package sshd

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

const (
	testingClientKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvziJnoiaaVyUGPnyqVC49XLzNRS+TPW63Nw4qovCG8lVbxKG
DIHC64tJrCDiZd0ppEhY+RQDGaPwrMInHnV8IwdS1wX22UTRuXA/oXmHcIxO2zmU
nrjFDlpKm2o+2Xd167ifdV9AiqNBtquO0M882RaGy99LbNPcl9ugAnxo5DVI1jES
l5vYqtiOAnRSvmJn2c+hkJfKXryH7hU4y+blDK5Vz44eSsC7bgG3ZbKfKGR9mlf2
ozVlzMNi2ACZ58vDBxn5WVLb1bPV1LHpicJ00fU3TDRnK3MkwvvAnqp78bNzi+ou
YIAwYSZ41iHNd596LQJchr1vs3Fo8qbgYaLY8wIDAQABAoIBAEJgQL0ME/Vw0mOd
F5OYVqu0vCF30trqDXQu6Wih3L5Cc+p7Vpau0Fds4STjwVK0o4jIKEJFpRHYa2m8
d1HGXFHYb/P9uQMQNXCWOzA0/EOgIJtOcH1sC9MAmpc6GRjps8AgNRHL/55gLyZW
hNuMpEWC4UWRfCAJpq/7554VS1+zWK0vy1GszikROjsZnopLTshMV+/7217tSk4O
1GY9ucNJX5iX3M83pmBOJX0ce8fqxeNnAdQIaAtp+ytm5TRzyaQtTjMlq0oqP8+7
Zx9aZKT11IpbOKBSIc6twRArlV1dT9kEI15zS9hfbWuvguB0zuhbhejS4wmZb9Tt
X8rGL4ECgYEA+MZcRzxpBKL+VNuQ4iSwF3RUYL1FIglJV7AM8UdM2hiNeiKidhD5
kmNXVf9C6XWg3OIHCno7HetBo0WZIPmOQMy4CDGC2bWEnQN+/bf3xsKzbECCLtH+
DALXSztihGGiY2zSoOCwTe7WZjGaF9s4C2rVkhsU/9di4qbapGTaWMECgYEAxMZD
c/sVTTT+/thdcLbBDhAfy6RMQwAy/1IPxNVR4C4O+l/rspbKxvV7JyaErP66g871
dBwrOGMfEsYoOOsUBFaj2/jJZdHvQj9jY/kdsfMBivHzkWEFte09NROOThbq+sgX
5bIPwS+IcVCgcA4We+aBv+rYKdvk05RJ8owPSrMCgYBjz4H6erxPxe1wsl8gvEOC
RYQNBCMWks9ARTwMGeU1o6AvnnG8GPdoyj6iHDYGYNFXjb/xbjUFvfupvCTB3B48
1WYIs4SiQHeiX2K1/PeGYVuHVSJmEo5w1zr1zi+qmVmDtoeTUFKsEeUnP0NpyuRj
gEuLwR3dv9bGxNb4GhaYgQKBgDNQCFL8TMe/ZCeMwIEeByXlqoTuKTznlmTiP15y
ylENcbZ0wP/nNqW/aggBkWOTYYvxsiw/FD42CupYZjDBjIy9EynPrKUyo5PA9+gg
FFBNMD/NbFii1lxkqytmGBvg+hG/kAvD7TvRa2ExR0UxR0e0Cm3Dje8MepV5+/aV
837lAoGBAPcvnrDFWKUy8dlrw05+9esiuZgCrCzZPw5xIxhrnRPcBOBl+QdpMscP
eWVutcVy5Frxl5tTf71WK/YhGPgWBt/CQz73Bf1+CX80CeApWWAqiAr240NED5a0
dBAFNBWp8IdHnQmdp9HKvxEXSK+RgOzPNLrpaRv+FPuiD6OtvhmD
-----END RSA PRIVATE KEY-----`
	testingClientFingerprint = `78:b9:21:20:1a:ed:e6:10:05:35:47:da:d4:1f:b6:73`
)

func sshTestingClientKey() (ssh.Signer, error) {
	return ssh.ParsePrivateKey([]byte(testingClientKey))
}

func TestFingerprint(t *testing.T) {
	key, _ := sshTestingClientKey()
	fp := fingerprint(key.PublicKey())
	if fp != testingClientFingerprint {
		t.Errorf("Expected fingerprint %s to match %s.", fp, testingClientFingerprint)
	}
}
