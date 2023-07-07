package bundlereader_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rancher/fleet/internal/bundlereader"
)

// fsNode represents a directory structure used to model `.fleetignore` test cases.
type fsNode struct {
	name string

	contents string   // if not a directory
	children []fsNode // non-empty only in case of a directory

	isDir bool
}

func TestGetContent(t *testing.T) {
	cases := []struct {
		name               string
		directoryStructure fsNode
		expectedFiles      map[string][]byte
	}{
		{
			name: "no .fleetignore",
			directoryStructure: fsNode{
				isDir: true,
				name:  "no-fleetignore",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml": []byte("foo"),
			},
		},
		{
			name: "empty .fleetignore",
			directoryStructure: fsNode{
				isDir: true,
				name:  "empty-fleetignore",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:     ".fleetignore",
						contents: "",
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml": []byte("foo"),
			},
		},
		{
			name: "ignore lines with leading # unless escaped",
			directoryStructure: fsNode{
				isDir: true,
				name:  "comments",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:     "#something_else.yaml",
						contents: "bar",
					},
					{
						name:     ".fleetignore",
						contents: "#something.yaml\n\\#something_else.yaml",
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml": []byte("foo"),
			},
		},
		{
			name: "simple .fleetignore",
			directoryStructure: fsNode{
				isDir: true,
				name:  "simple-fleetignore",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:     "something_else.yaml",
						contents: "bar",
					},
					{
						name:     ".fleetignore",
						contents: "not_here.txt\nsomething.yaml",
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something_else.yaml": []byte("bar"),
			},
		},
		{
			name: "glob syntax",
			directoryStructure: fsNode{
				isDir: true,
				name:  "glob-syntax",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:     ".fleetignore",
						contents: "something*",
					},
				},
			},
			expectedFiles: map[string][]byte{},
		},
		{
			name: "ignore trailing spaces unless escaped",
			directoryStructure: fsNode{
				isDir: true,
				name:  "trim-space",
				children: []fsNode{
					{
						name:     "something.yaml ",
						contents: "foo",
					},
					{
						name:     "something_else.yaml  ",
						contents: "bar",
					},
					{
						name:     ".fleetignore",
						contents: "something_else.yaml\\ \\ \nsomething.yaml ",
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml ": []byte("foo"),
			},
		},
		{
			name: "ignore directories",
			directoryStructure: fsNode{
				isDir: true,
				name:  "ignore-directories",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:     ".fleetignore",
						contents: "subdir",
					},
					{
						name:  "subdir",
						isDir: true,
						children: []fsNode{
							{
								name:     "in_dir.yaml",
								contents: "baz",
							},
						},
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml": []byte("foo"),
			},
		},
		{
			name: "ignore file multiple levels below .fleetignore",
			directoryStructure: fsNode{
				isDir: true,
				name:  "ignore-file-multiple-levels",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:     ".fleetignore",
						contents: "in_dir.yaml",
					},
					{
						name:  "subdir",
						isDir: true,
						children: []fsNode{
							{
								name:  "subsubdir",
								isDir: true,
								children: []fsNode{
									{
										name:     "in_dir.yaml",
										contents: "bar",
									},
								},
							},
						},
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml": []byte("foo"),
			},
		},
		{
			name: ".fleetignore files in neighbour dirs do not interfere with one another",
			directoryStructure: fsNode{
				isDir: true,
				name:  "multiple-files-same-level",
				children: []fsNode{
					{
						name:     "something.yaml",
						contents: "foo",
					},
					{
						name:  "subdir1",
						isDir: true,
						children: []fsNode{
							{
								name:     "in_dir.yaml",
								contents: "from dir 1",
							},
							{
								name:     ".fleetignore",
								contents: "in_dir.yaml",
							},
						},
					},
					{
						name:  "subdir2",
						isDir: true,
						children: []fsNode{
							{
								name:     "in_dir.yaml",
								contents: "from dir 2",
							},
						},
					},
				},
			},
			expectedFiles: map[string][]byte{
				"something.yaml":      []byte("foo"),
				"subdir2/in_dir.yaml": []byte("from dir 2"),
			},
		},
	}

	base, err := os.MkdirTemp("", "test-fleet")
	require.NoError(t, err)

	defer os.RemoveAll(base)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			root := createDirStruct(t, base, c.directoryStructure)

			files, err := bundlereader.GetContent(context.Background(), root, root, "", bundlereader.Auth{})
			assert.NoError(t, err)

			assert.Equal(t, len(c.expectedFiles), len(files))
			for k, v := range c.expectedFiles {
				assert.Equal(t, v, files[k])
			}
		})
	}
}

// createDirStruct generates and populates a directory structure which root is node, placing it at basePath.
func createDirStruct(t *testing.T, basePath string, node fsNode) string {
	path := filepath.Join(basePath, node.name)

	if !node.isDir {
		f, err := os.Create(path)
		require.NoError(t, err)

		_, err = io.WriteString(f, node.contents)
		require.NoError(t, err)

		return ""
	}

	err := os.Mkdir(path, 0777)
	require.NoError(t, err)

	for _, c := range node.children {
		createDirStruct(t, path, c)
	}

	return path
}
