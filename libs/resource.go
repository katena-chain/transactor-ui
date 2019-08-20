package libs

import (
    "fmt"
    "os"

    "fyne.io/fyne"

    "github.com/katena-chain/transactor-ui/assets"
)

func MakeImageResource(name string, path string) (*fyne.StaticResource, error) {
    // Builds a fyne.Resource out of an image

    resultBytes, err := assets.Asset(path)
    if err != nil {
        return nil, err
    }

    resultResource := fyne.NewStaticResource(name, resultBytes)
    return resultResource, nil
}

func CheckIcon(err error) {
    if err != nil {
        fmt.Println("Error encountered opening an icon", err.Error())
        os.Exit(1)
    }
}
