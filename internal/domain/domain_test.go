package domain_test

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/user"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/mckern/spiry/internal/domain"
	"github.com/stretchr/testify/assert"
)

var localWhoisServer net.Addr

func unprivilegedUser() bool {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	return user.Uid != "0"
}

func handleConnection(c net.Conn) {
	log.Println("waiting for input...")
	// NewReader should handle crlf line endings for us
	netData, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	name := strings.TrimSpace(string(netData))

	log.Printf("reading fake whois data for domain %v\n", name)
	file, err := os.Open(path.Join("fixtures", name+".whois"))
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, err := c.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	c.Close()
}

func startWhoisServer(l net.Listener) {
	log.Printf("spinning up fake whois server on %v\n", l.Addr().String())

	for {
		log.Println("now accepting connections...")
		c, _ := l.Accept()
		go handleConnection(c)
	}
}

func stopWhoisServer(l net.Listener) {
	log.Printf("stopping fake whois server on %v\n", l.Addr().String())
	l.Close()
}

func TestMain(m *testing.M) {
	if !unprivilegedUser() {
		listener, err := net.Listen("tcp", "127.0.0.1:43")
		if err != nil {
			log.Fatal(err)
		}

		localWhoisServer = listener.Addr()

		defer stopWhoisServer(listener)
		go startWhoisServer(listener)
	}

	os.Exit(m.Run())
}

func TestDomainRoot(t *testing.T) {
	d := domain.New("subdomain.mckern.sh")
	root, _ := d.Root()
	assert.NotEqual(t, root, d.Name, "the root domain should be parsed from a FQDN")
}

func TestDomainExpiry(t *testing.T) {
	if unprivilegedUser() {
		t.Skip("Skipping testing in unprivileged environment")
	}

	d := domain.New("mckern.sh")
	d.WhoisServer = localWhoisServer.String()
	val, _ := d.Expiry()

	assert.IsType(t, time.Time{}, val, "an expiration date should be a valid time.Time instance")
	assert.NotNil(t, val, "a domain should have a real expiration date")
}
