package main

import (
	"fmt"
	"log"
	"memcached-dump/client"
	"net"
	"os"
)

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	if len(os.Args) < 2 {
		printDefault()
		return
	}
	switch os.Args[1] {
	case "version":
		if len(os.Args) < 3 {
			printDefault()
			return
		}
		address := os.Args[2]
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("invalid address format " + address)
			return
		}
		cli, err := client.NewClient(addr)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer cli.Close()
		ver, err := cli.Version()
		if err != nil {
			log.Fatalln(err)
			return
		}
		log.Println("memcached version", ver)
		break
	case "keys":
		if len(os.Args) < 3 {
			printDefault()
			return
		}
		address := os.Args[2]
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("invalid address format " + address)
			return
		}
		cli, err := client.NewClient(addr)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer cli.Close()

		keys, err := cli.GetKeys()
		if err != nil {
			log.Fatalln(err)
			return
		}
		if len(keys) > 0 {
			i := 0
			fmt.Println("----------------------------------------")
			for name, _ := range keys {
				i++
				fmt.Println(i, name)
			}
			fmt.Println("----------------------------------------")
		}
		log.Println("Find keys", len(keys))
		break
	case "count":
		if len(os.Args) < 3 {
			printDefault()
			return
		}
		address := os.Args[2]
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("invalid address format " + address)
			return
		}
		cli, err := client.NewClient(addr)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer cli.Close()

		keys, err := cli.GetKeys()
		if err != nil {
			log.Fatalln(err)
			return
		}
		log.Println("Find keys", len(keys))
		break
	case "store":
		if len(os.Args) < 3 {
			printDefault()
			return
		}
		address := os.Args[2]
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("invalid address format " + address)
			return
		}
		cli, err := client.NewClient(addr)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer cli.Close()

		n, err := cli.Store()
		if err != nil {
			log.Fatalln(err)
			return
		}
		if n <= 0 {
			log.Fatalln("Get empty data and nothing store")
			return
		}
		break
	case "dump":
		if len(os.Args) < 4 {
			printDefault()
			return
		}
		address := os.Args[2]
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("invalid address format " + address)
			return
		}
		address1 := os.Args[3] // 目标memcached
		if address == address1 {
			log.Fatalln("destination address is same with " + address)
			return
		}
		addr1, err := net.ResolveTCPAddr("tcp", address1)
		if err != nil {
			log.Fatalln("invalid address format " + address1)
			return
		}
		cli, err := client.NewClient(addr)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer cli.Close()

		n, err := cli.DumpTo(addr1)
		if err != nil {
			log.Fatalln(err)
			return
		}
		if n <= 0 {
			log.Println("Get empty data and nothing change")
			return
		}
		break
	default:
		printDefault()
		break
	}
}

func printDefault() {
	fmt.Println("Usage :")
	fmt.Println("   version <IP:PORT> show the memcached version")
	fmt.Println("   count <IP:PORT>   count the keys")
	fmt.Println("   keys <IP:PORT>    list all keys")
	fmt.Println("   store <IP:PORT>   store data to local file")
	fmt.Println("   dump <IP:PORT> <IP:PORT>  dump all data to another memcached")
}
