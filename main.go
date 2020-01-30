package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/matrix-org/gomatrix"
)

func sendMessage(c *gomatrix.Client, roomID, message string) error {
	_, err := c.UserTyping(roomID, true, 3)
	if err != nil {
		return err
	}

	c.SendText(roomID, message)

	_, err = c.UserTyping(roomID, false, 0)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var store, err = NewStore("mcchunkie.db")
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	var username, password, userID, accessToken, server string
	var setup bool

	flag.StringVar(&username, "user", "", "username to connect to matrix server with")
	flag.StringVar(&server, "server", "", "matrix server")
	flag.BoolVar(&setup, "s", false, "setup account")

	flag.Parse()

	if server == "" {
		server, err = store.get("config", "server")
		if server == "" {
			log.Fatalln("please specify a server")
		}

	} else {
		store.set("config", "server", server)
	}

	log.Printf("connecting to %s\n", server)

	cli, err := gomatrix.NewClient(
		server,
		"",
		"",
	)

	if setup {
		log.Println("requesting access token")
		password, err = prompt(fmt.Sprintf("Password for '%s': ", username))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println()

		log.Printf("'%s' : '%s'\n", username, password)

		resp, err := cli.Login(&gomatrix.ReqLogin{
			Type:     "m.login.password",
			User:     username,
			Password: password,
		})
		if err != nil {
			log.Fatalln(err)
		}

		store.set("account", "username", username)
		store.set("account", "access_token", resp.AccessToken)
		store.set("account", "user_id", resp.UserID)
	} else {
		username, _ = store.get("account", "username")
		accessToken, _ = store.get("account", "access_token")
		userID, _ = store.get("account", "user_id")
	}

	cli.SetCredentials(userID, accessToken)
	cli.Store = store
	syncer := gomatrix.NewDefaultSyncer(username, store)
	cli.Client = http.DefaultClient
	cli.Syncer = syncer

	/*
		if _, err := cli.JoinRoom("!tmCVBJAeuKjCfihUjb:cobryce.com", "", nil); err != nil {
			log.Fatalln(err)
		}
		if _, err := cli.JoinRoom("!sFPUeGfHqjiItcjNIN:matrix.org", "", nil); err != nil {
			log.Fatalln(err)
		}
		if _, err := cli.JoinRoom("!ALCZnrYadLGSySIFZr:matrix.org", "", nil); err != nil {
			log.Fatalln(err)
		}
	*/
	if _, err := cli.JoinRoom("!LTxJpLHtShMVmlpwmZ:tapenet.org", "", nil); err != nil {
		log.Fatalln(err)
	}

	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Sender == username {
			return
		}

		if mtype, ok := ev.MessageType(); ok {
			switch mtype {
			case "m.text":
				if post, ok := ev.Body(); ok {
					log.Printf("%s: '%s'", ev.Sender, post)
				}
			}
		}
	})

	//cli.SendText("!tmCVBJAeuKjCfihUjb:cobryce.com", "Butts")
	sendMessage(cli, "!LTxJpLHtShMVmlpwmZ:tapenet.org", "Typing hi!")
	sendMessage(cli, "!tmCVBJAeuKjCfihUjb:cobryce.com", "Butts")

	avatar := "https://deftly.net/mcchunkie.png"
	aurl, err := cli.GetAvatarURL()

	if aurl != avatar {
		log.Printf("Setting avatar to: '%s'", avatar)
		err = cli.SetAvatarURL(avatar)
		if err != nil {
			fmt.Println("Unable to set avatar: ", err)
		}
	} else {
		log.Printf("avatar already set")
	}

	for {
		log.Println("syncing..")
		if err := cli.Sync(); err != nil {
			fmt.Println("Sync() returned ", err)
		}

		time.Sleep(1 * time.Second)
	}
}
