package sshd

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

const (
	testingClientKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEArjUDq6/7ljVzoa7unbdSRMNIwfFd7S0YM931w7YstZXFvnuN
eavoAxDkL0mdWxV0Pi6f+FFi31oY3YHUBaWdvkZHXCY9L3zWRKz00SRNnyeQG8tO
GGhvhvgC6iGIE6A9IJlLxDm6scylp6JaN27P4CUNXy8gT0GnxvdwMGujgbMbPU2x
XAC5JnT1SP++7wJeZDwM1TXGn386EuTBN1epIsIvmsriW7eOXvOOf+7eQy/RbCJo
51rmF9kMnUqRvY5Be0Ur4D842JiFGRIVyon/IGoBD2d/JOPi36npJxz47LdTWsL0
dLREL009U8ttC00QuKoScHq4G39Jw8NmXQhhXwIDAQABAoIBAEhkW2wjK2dWOwD7
UslTfup4RGnjxWZkEOSs3g5ATAABhzUK3tWq7DUp9cj4zF0nYzDb6zojh/TM2fxi
kRrvoceKKOlQMqjjNZ9ASFQIxADZTfde2sslywLJWVy2JngRZJWBXozieISeSFCL
FPZoJBY/D3l4efK1k+UIuiRE9qNUfKq/XlZuh57nXJ/FoNdPeCh68Q0EcxlDiEmZ
cqpgKmRSnQDZBj+gaVDy6LQQ353ZlLkArPvaJFHo4pjUEsmvmG6GWkh2yR3MTdIA
LiN7MS8bk6zcL/1grQ6/R5we12A9V+XVvou4VPAwRacv74nIbbU8Y4UztrGentsN
+Bq5B0ECgYEA4I2ABmWdWluUrpUsXZEIgCVu3ee1FQmrqXUSeanGS+uaXXBJGbV5
kxvZqRaZ+mHxtOwAhmmHl/2M+hKc64m9018TX4BcDyXA/6HND1TJxxStYw92lCMG
tBT1NPoNKnHH9goJgLO24y3qu2Aax7FuF7YPLR8r0agfRUiWg9B2yQ8CgYEAxpqP
HCP4V1peElVtqppqTBz4S9Li9dnR+JuMkQZxt2nUuy29hR0SJlbIXRohjN3UHIOK
c1wMy0EIXmcLxApRpEDKcm1zcDF3LL7l0hFFX4+JYTcVNn5VLeUUQtFGmDaSQ/Y5
dT59kfSu8zVX9ZecUAUV9CLJC4MF7F0gCJzWwrECgYBKcp9ff5ELxBEnUI3E97C5
y69WItwGfY5MQGQ/sensgdBL6k5SF7iW7UTcqoGiYZahRR1nctVhrs5umn0sGh61
VXA22XesDfhOyHYT/yhmuJRDo3zM4E/4pHondj+nMtH44JsF8I9SAocwWEyIqGq3
scSWUR9WA0da0RYV3aeEQQKBgC9mCcukhguLBLKJcu/phH7/1v55qTMVtjgIH6cp
C5DDkELP6tBPHNrLkWwu5VzyQEJB3pQjnuYPckjdfQBfmhaCZA6lMozPMWsbcEwP
VSg2YIo0FDr6MagPaSN9QMTpGUVhCVuC+4MPC4X98C0r7uFmJVQrzSGTNqGvpAqK
K/MxAoGAX6SsKbBlANkAm4bhHRj7xse31n8mUT/ewp4h6TmknL4aSoPH+PqgDQQ5
WauSC6B2gAKgYogsDa+Ij8ck2NFFlPyeCuW88FOUXXBbOTj+S2dscJ85OIiZX7MV
hnpuSad2mCqNaqwU+/9ANrycBpaQtyHBspAYuO3/UUbilmJKgLo=
-----END RSA PRIVATE KEY-----`
	testingClientFingerprint = `fa:61:1a:1f:45:6a:fa:32:5f:18:c4:4b:a5:b3:99:a3`
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
