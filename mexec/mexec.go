package mexec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type Bin struct {
	Data *[]byte
	Fd   *os.File
}

func (b *Bin) Close() error {
	if b.Fd != nil {
		err := b.Fd.Close()
		b.Fd = nil
		return err
	}
	return nil
}

func (b *Bin) Open() error {
	err := b.Close()
	if err != nil {
		return err
	}
	a := make([]byte, 1)
	r, _, e := syscall.Syscall(319, uintptr(unsafe.Pointer(&a[0])), uintptr(0x1), 0)
	if e != 0 {
		return errors.New("memfd_create failed")
	}
	b.Fd = os.NewFile(r, fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), int(r)))
	if _, err = b.Fd.Write(*b.Data); err != nil {
		_ = b.Fd.Close()
		return err
	}
	return nil
}

func (b *Bin) Command(args ...string) *exec.Cmd {
	return exec.Command(b.Fd.Name(), args...)
}

func NewBin(data *[]byte) (*Bin, error) {
	if len(*data) == 0 {
		return nil, errors.New("binary data is empty")
	}
	bin := &Bin{Data: data}
	if err := bin.Open(); err != nil {
		return nil, err
	}
	return bin, nil
}

func OpenBin(file string) (*Bin, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary file %s: %w", file, err)
	}
	if len(b) == 0 {
		return nil, errors.New("binary file is empty")
	}
	return NewBin(&b)
}
