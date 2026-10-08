//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// gen.go builds imagedumper.jar, a self-contained executable jar bundling
// ImageDumper.class and the vendored HSQLDB driver, so that the extraction
// command needs nothing but `java -jar`.

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s %v: %v\n", name, args, err)
		os.Exit(1)
	}
}

func main() {
	staging, err := os.MkdirTemp("", "imagedumper-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create staging dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(staging)

	src, err := filepath.Abs("../../../java/ImageDumper.java")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	depJar, err := filepath.Abs("../../../third_party/hsqldb-1.8.1.3/hsqldb.jar")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 1. Compile ImageDumper into the staging dir.
	run("javac", "--release", "8", "-nowarn", "-d", staging, src)

	// 2. Unpack the HSQLDB driver classes into the staging dir.
	unpack := exec.Command("jar", "-xf", depJar)
	unpack.Dir = staging
	unpack.Stdout = os.Stdout
	unpack.Stderr = os.Stderr
	if err := unpack.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to unpack %s: %v\n", depJar, err)
		os.Exit(1)
	}
	// The vendored jar's own manifest would shadow ours; discard it.
	os.RemoveAll(filepath.Join(staging, "META-INF"))

	// 3. Write a manifest making the jar directly executable.
	manifest := filepath.Join(staging, "manifest.txt")
	if err := os.WriteFile(manifest, []byte("Main-Class: ImageDumper\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write manifest: %v\n", err)
		os.Exit(1)
	}

	// 4. Package everything as imagedumper.jar.
	run("jar", "-cfm", "imagedumper.jar", manifest, "-C", staging, ".")

	fmt.Println("Successfully built imagedumper.jar")
}
