package main

import (
      "github.com/czz/oblivion/core/session"
            "github.com/czz/oblivion/core/tui"
            "fmt"
)


func main() {
  t := tui.NewTui()

  fmt.Println(t.Yellow(`
  ▄██████▄  ▀█████████▄   ▄█        ▄█   ▄█    █▄   ▄█   ▄██████▄  ███▄▄▄▄
 ███    ███   ███    ███ ███       ███  ███    ███ ███  ███    ███ ███▀▀▀██▄
 ███    ███   ███    ███ ███       ███▌ ███    ███ ███▌ ███    ███ ███   ███
 ███    ███  ▄███▄▄▄██▀  ███       ███▌ ███    ███ ███▌ ███    ███ ███   ███
 ███    ███ ▀▀███▀▀▀██▄  ███       ███▌ ███    ███ ███▌ ███    ███ ███   ███
 ███    ███   ███    ██▄ ███       ███  ███    ███ ███  ███    ███ ███   ███
 ███    ███   ███    ███ ███▌    ▄ ███  ███    ███ ███  ███    ███ ███   ███
  ▀██████▀  ▄█████████▀  █████▄▄██ █▀    ▀██████▀  █▀    ▀██████▀   ▀█   █▀
                         ▀
`));

    fmt.Println("Made with ❤️  by czz78")

    // Start a new session
    s := session.NewSession()
    // Start the session
    s.Start()
    // Enter the input reading loop
    for s.Active {
        s.ReadlineLoop()
    }

}
