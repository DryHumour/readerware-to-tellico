//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	// 1. Run javac to compile ImageDumper.java
	cmd := exec.Command("javac", "--release", "8", "-nowarn", "-d", ".", "../../../java/ImageDumper.java")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: failed to compile ImageDumper.java: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully compiled ImageDumper.java")

	// 2. Copy hsqldb.jar
	srcJar := "../../../third_party/hsqldb-1.8.1.3/hsqldb.jar"
	dstJar := "hsqldb.jar"

	if err := copyFile(srcJar, dstJar); err != nil {
		fmt.Printf("Error: failed to copy hsqldb.jar: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully copied hsqldb.jar")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
