package gitreceive

import (
	"testing"
)

type checkCase struct {
	podTick     int
	podWait     int
	storageTick int
	storageWait int
}

func TestCheckDurations(t *testing.T) {
	cases := map[checkCase]checkCase{
		checkCase{100, 300000, 500, 300000}:    {100, 300000, 500, 300000},
		checkCase{0, 300000, 500, 300000}:      {100, 300000, 500, 300000},
		checkCase{100, 300000, 0, 300000}:      {100, 300000, 500, 300000},
		checkCase{300000, 300000, 500, 300000}: {100, 300000, 500, 300000},
		checkCase{100, 300000, 300000, 300000}: {100, 300000, 500, 300000},
	}

	var cnf Config
	for tCase, eCase := range cases {
		cnf = Config{
			BuilderPodTickDurationMSec:    tCase.podTick,
			BuilderPodWaitDurationMSec:    tCase.podWait,
			ObjectStorageTickDurationMSec: tCase.storageTick,
			ObjectStorageWaitDurationMSec: tCase.storageWait,
		}
		cnf.CheckDurations()

		if cnf.BuilderPodTickDurationMSec != eCase.podTick {
			t.Fatalf("expected %v but %v was returned (%v)", eCase.podTick, cnf.BuilderPodTickDurationMSec, tCase)
		}
		if cnf.BuilderPodWaitDurationMSec != eCase.podWait {
			t.Fatalf("expected %v but %v was returned (%v)", eCase.podWait, cnf.BuilderPodWaitDurationMSec, tCase)
		}
		if cnf.ObjectStorageTickDurationMSec != eCase.storageTick {
			t.Fatalf("expected %v but %v was returned (%v)", eCase.storageTick, cnf.ObjectStorageTickDurationMSec, tCase)
		}
		if cnf.ObjectStorageWaitDurationMSec != eCase.storageWait {
			t.Fatalf("expected %v but %v was returned (%v)", eCase.storageWait, cnf.ObjectStorageWaitDurationMSec, tCase)
		}
	}
}
