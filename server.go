package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	bolt "go.etcd.io/bbolt"
)

const KeyLen = 32
const Endpoint = "192.168.56.102"


db, err := bolt.Open("my.db", 0600, nil)
if err != nil {
	log.Fatal(err)
}

type Config struct {
	ServerIp         string
	ClientIp         string
	ServerPublicKey  string
	ClientPrivateKey string
}

func main() {
	is_root := os.Getenv("SUDO_USER")

	if len(os.Args[1:]) != 1 {
		fmt.Println("Please state the device you want to control.")
		os.Exit(1)
	}

	if len(is_root) < 1 {
		fmt.Println("Please run this program as root.")
		os.Exit(1)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	http.HandleFunc("/add-device", AddDevice)

	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func gen_ip(cidr string) (ip net.IP) {
	_, network, _ := net.ParseCIDR(cidr)
	random_base := make([]byte, 4)
	random_ip := make(net.IP, net.IPv4len)

	binary.LittleEndian.PutUint32(random_base, rand.Uint32())

	for i, v := range random_base {
		random_ip[i] = network.IP[i] | (v & ^network.Mask[i])
	}

	return net.IP(random_ip)
}

func AddDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	cli, _ := wgctrl.New()
	device, error := cli.Device(os.Args[1])
	priv_key, _ := wgtypes.GeneratePrivateKey()
	// pub_key := wgtypes.Key.PublicKey(priv_key)
	ip_address := gen_ip("172.16.0.0/12")

	if error != nil {
		fmt.Println("Could not access device.")
		os.Exit(1)
	}

	for _, peer := range device.Peers {
		fmt.Println(peer.PublicKey)
	}

	// new_config := wgtypes.Config{
	// 	ReplacePeers: true,
	// 	Peers: []wgtypes.PeerConfig{
	// 		{
	// 			PublicKey:  pub_key,
	// 			AllowedIPs: []net.IPNet{*ip_address},
	// 		},
	// 	},
	// }

	// cli.ConfigureDevice(os.Args[1], new_config)

	config := Config{
		ServerIp:         Endpoint,
		ServerPublicKey:  device.PublicKey.String(),
		ClientPrivateKey: priv_key.String(),
		ClientIp:         ip_address.String(),
	}
	encoder.Encode(config)
}
