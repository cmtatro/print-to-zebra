package main

import (
    "io"
    "io/ioutil"
    "net/http"
    "fmt"
    "log"
    "os"
    "os/user"
    "os/exec"
    "bytes"
    "strings"
    "golang.org/x/sys/windows/registry"
    "crypto/sha256"
    "encoding/hex"
)
func hold() {
    var input string
    fmt.Scanln(&input)
}

func check(e error) {
    if e != nil {
        log.Println(e)
        fmt.Print("\r\nThere seems to have been an error. Please contact IT support")
        var input string
        fmt.Scanln(&input)
//        panic(e)
    }
}

func main() {
    args := os.Args[1:]

    if len(args) == 0 || args[0] == "install" {
        fmt.Println("Installing registry keys")
        install()

        fmt.Print("Installation Complete. Press any Enter to exit...")
        var input string
        fmt.Scanln(&input)

        os.Exit(0)
    }

    if len(args) == 0 || args[0] == "hashKey" {
        s:=getHash()

        fmt.Print("Secret Hash is: ")
        fmt.Printf("%s", s)
        fmt.Print("\n\nPress enter to exit...")
        var input string
        fmt.Scanln(&input)

        os.Exit(0)
    }

    data := args[0]
    domain := "someDomain"

    if !strings.Contains(data, domain) {
        log.Fatal("Specific domain is required to use this function")
    }

    // Download file from internet if we need to.
    if (strings.Contains(data, "zebra://")) {
        path := strings.TrimRight(strings.TrimLeft(data, "zebra://"), "/")
        // Create an array of strings to join below. Leave off ending slash as the seperator will add it.
        pathAndProtocol:= []string{"https://", path, "?token=", getHash()}
        fmt.Print("Getting label from: \n")
        fmt.Print(strings.Join(pathAndProtocol, ""))
        fmt.Print("\n\n")

        dlErr:=DownloadFile(`C:\ZebraPrintLabel\ShippingLabel.EPL`, strings.Join(pathAndProtocol, ""))
        check(dlErr)
    } else {

        // Load file from local storage
        filerc, err := os.Open(data)
        if err != nil{
            log.Fatal(err)
        }
        defer filerc.Close()

        buf := new(bytes.Buffer)
        buf.ReadFrom(filerc)
        contents := buf.String()

        d1 := []byte(contents)

        ioerr := ioutil.WriteFile(`C:\ZebraPrintLabel\ShippingLabel.EPL`, d1, 0644)
        check(ioerr)
    }

    checkContents()

    c := exec.Command(`C:\ZebraPrintLabel\print_label.bat`, `C:\ZebraPrintLabel\ShippingLabel.EPL`).Run()
    fmt.Print("Printed label...\n")
    //var input string
    //fmt.Scanln(&input)
    check(c)

    // Print file from internet
    os.Exit(0)
}

func getHash() (string) {
    hashSalt := "something2343948031432"
    currentUser, hErr := user.Current()
    // Debug current user
    // fmt.Print(currentUser.Username)
    // var input string
    // fmt.Scanln(&input)
    check(hErr)
    fmt.Print(currentUser.Username)
    s := []string{currentUser.Username, hashSalt}

    // The pattern for generating a hash is `sha256.New()`,
    // `sha256.Write(bytes)`, then `sha256.Sum([]byte{})`.
    // Here we start with a new hash.
    h := sha256.New()

    // `Write` expects bytes. If you have a string `s`,
    // use `[]byte(s)` to coerce it to bytes.
    h.Write([]byte(strings.Join(s, "")))

    // This gets the finalized hash result as a byte
    // slice. The argument to `Sum` can be used to append
    // to an existing byte slice: it usually isn't needed.
    bs := hex.EncodeToString(h.Sum(nil))

    return bs
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

    // Create the file
    fmt.Print("Creating label... \n")
    out, err := os.Create(filepath)
    if err != nil {
        fmt.Print("Failed")
        var input string
        fmt.Scanln(&input)
        return err
    }
    defer out.Close()

    // Get the data
    fmt.Print("Getting label...\n")
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    // Verify we weren't redirected

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return err
    }

    return nil
}

func install() {
    zexe := exec.Command(`C:\ZebraPrintLabel\Zebra_7.4.3.exe`).Run()
    check(zexe)

    k, opBool, err := registry.CreateKey(registry.CLASSES_ROOT, `zebra`, registry.QUERY_VALUE|registry.SET_VALUE|registry.CREATE_SUB_KEY)
    c, c_opBool, c_err := registry.CreateKey(k, `shell\open\command`, registry.QUERY_VALUE|registry.SET_VALUE)

    if err != nil {
        fmt.Println("err")
        fmt.Println(opBool)
        fmt.Println(c_opBool)
        log.Fatal(err)
    }
    if err != nil {
        fmt.Println("c_err")
        log.Fatal(c_err)
    }
    if err := k.SetStringValue("URL Protocol", ""); err != nil {
        fmt.Println("url protocol")
        log.Fatal(err)
    }
    if err := k.SetStringValue("", "\"URL:Zebra Protocol\""); err != nil {
        fmt.Println("Default")
        log.Fatal(err)
    }
    if err := c.SetStringValue("", `"C:\ZebraPrintLabel\print-to-zebra.exe" "%1"`); err != nil {
        fmt.Println("cmd")
        log.Fatal(err)
    }
    if err := k.Close(); err != nil {
        log.Fatal(err)
    }
}

func checkContents() {
    f, err := os.Open(`C:\ZebraPrintLabel\ShippingLabel.EPL`)
    check(err)
    b1 := make([]byte, 4)
    n1, err := f.Read(b1)
    check(err)

    fmt.Printf("%d bytes: %s\n", n1, string(b1))
    if (string(b1) != "EPL2") {
        fmt.Print("File downloaded was not EPL2. Aborting print...\n")
        fmt.Print("Please run \"C:\\ZebraPrintLabel\\print-to-zebra.exe hashKey\" and provide the Key to ITSupport.")
        hold()
    }
}
