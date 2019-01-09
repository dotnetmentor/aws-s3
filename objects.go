package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// A Object provides details of an S3 object
type Object struct {
	Bucket       string
	Key          string
	Folder       string
	File         string
	LastModified time.Time
	Age          time.Duration
}

// SortalbeObjects provides sorting of S3 objects
type SortalbeObjects []Object

func (s SortalbeObjects) Len() int      { return len(s) }
func (s SortalbeObjects) Swap(a, b int) { s[a], s[b] = s[b], s[a] }
func (s SortalbeObjects) Less(a, b int) bool {
	if s[a].Key < s[b].Key {
		return true
	}

	return false
}

func sortObjects(objs []Object) {
	s := SortalbeObjects(objs)
	sort.Sort(s)
}

func printObjects(objs []Object, includeFiles bool, includeDirectories bool) {
	folder := ""
	for _, obj := range objs {
		if obj.Folder != folder && includeDirectories {
			folder = obj.Folder
			fmt.Printf("  %s\n", folder)
		}

		if includeFiles {
			fmt.Printf("    %s (modified=%v age=%vh)\n", obj.File, obj.LastModified, math.Round(obj.Age.Hours()))
		}
	}
}

func countFolders(objs []Object) int {
	folder := ""
	folders := 0
	for _, obj := range objs {
		if obj.Folder != folder {
			folder = obj.Folder
			folders++
		}
	}
	return folders
}
