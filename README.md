# sftp

This repository provides a small client for interacting with remote SFTP
servers. This is primarily for uploading inventory to third-party listing sites,
though it can easily be used for other tasks as well.

> Note: This package is not currently suitable for remote servers that use key
> based authentication instead of username/password.

## Usage

Usage is pretty simple, and only requires a minimum of input parameters. Error
handling has been omitted in the following listing for brevity.

```go
func main() {
    ctx := context.Background()

    // Get your username, password, and address (including port) from whatever
    // mechanism you prefer (though of course don't hardcode them in anything
    // being sent to version control or other public location).
    client, _ := sftp.NewClient(sftp.Config{
        User:     "",
        Password: "",
        Addr:     "",
    })
    // If your client is not intended to be alive for the full run of your
    // application, remember to close it to prevent keeping open handles.
    defer client.Close()

    // There are three save methods which take various inputs: an io.Reader, a
    // string, or a byte slice. The latter two are convenience wrappers around
    // the bare Save method, so we show those here, but know that anything that
    // can be made into an io.Reader can be uploaded.
    _ = client.SaveString(ctx, "test2.txt", "This is a test.")
    _ = client.SaveBytes(ctx, "test3.txt", []byte("This is a test."))

    // To get a file listing, you can call ReadDir passing in the path to the
    // directory in question. You get back a slice of os.FileInfo that can be
    // used just as you would with a local file.
    files, _ := client.ReadDir(ctx, ".")

    // To remove a file from remote (assuming it's allowed), just call Remove
    // with the path.
    _ = client.Remove(ctx, "test3.txt"); err != nil {
}
```
