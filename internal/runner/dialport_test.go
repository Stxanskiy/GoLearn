package runner

import (
	"bufio"
	"net"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// The relay has to survive silence. A VNC server sends nothing at all while the
// picture is still, and an idle timeout on its socket would take the relay down
// with it — the window appeared to close by itself after eleven seconds.
//
// This runs the real script against a real listener instead of matching the
// source text, because the bug was in what Python does with the argument, not
// in what the script says.
func TestPortRelaySurvivesSilence(t *testing.T) {
	if testing.Short() {
		t.Skip("ждёт дольше, чем длится таймаут, который проверяет")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("нужен python3")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	cmd := exec.Command("python3", "-c", relayScript(port))
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = in.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	srv, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	r := bufio.NewReader(out)
	if _, err := srv.Write([]byte("первый кадр\n")); err != nil {
		t.Fatal(err)
	}
	if line, err := r.ReadString('\n'); err != nil || line != "первый кадр\n" {
		t.Fatalf("первый кадр не дошёл: %q, %v", line, err)
	}

	// Дольше таймаута, который раньше стоял на сокете.
	time.Sleep(12 * time.Second)

	if _, err := srv.Write([]byte("кадр после тишины\n")); err != nil {
		t.Fatalf("релей закрыл соединение, пока ничего не происходило: %v", err)
	}
	line, err := r.ReadString('\n')
	if err != nil || line != "кадр после тишины\n" {
		t.Fatalf("после тишины релей умер: %q, %v", line, err)
	}
}
