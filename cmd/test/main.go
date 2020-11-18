package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/frazerhub/sftp"
)

var (
	user = os.Getenv("MYNEXTRIDE_USER")
	pass = os.Getenv("MYNEXTRIDE_PASSWORD")
	addr = os.Getenv("MYNEXTRIDE_ADDRESS")
)

func main() {
	ctx := context.Background()

	client, err := sftp.NewClient(sftp.Config{
		User:     user,
		Password: pass,
		Addr:     addr,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	f, err := os.Open("./cmd/test/test1.txt")
	if err != nil {
		panic(err)
	}

	fmt.Println("Saving file from io.Reader.")
	if err := client.Save(ctx, "test1.txt", f); err != nil {
		fmt.Println("ERROR: reader test:", err)
	}

	fmt.Println("Saving file from string.")
	if err := client.SaveString(ctx, "test2.txt", "This is a test."); err != nil {
		fmt.Println("ERROR: string test:", err)
	}

	fmt.Println("Saving file from bytes.")
	if err := client.SaveBytes(ctx, "test3.txt", []byte("This is a test.")); err != nil {
		fmt.Println("ERROR: bytes test:", err)
	}
	f.Close()
	fmt.Println()

	files, err := client.ReadDir(ctx, ".")
	if err != nil {
		panic(err)
	}
	fmt.Println("Found files:")
	for _, fi := range files {
		fmt.Printf("%s\n%s\n", fi.Name(), strings.Repeat("-", len(fi.Name())))
		f, err := client.Open(ctx, fi.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}
		io.Copy(os.Stderr, f)
		f.Close()
		fmt.Println()
		fmt.Println()
	}

	for _, fi := range files {
		fmt.Printf("Deleting %s.\n", fi.Name())
		if err := client.Remove(ctx, fi.Name()); err != nil {
			fmt.Println(err)
			continue
		}
	}
}
