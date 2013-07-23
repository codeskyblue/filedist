package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Go is a basic promise implementation: it wraps calls a function in a goroutine,
// and returns a channel which will later return the function's return value.
func Go(f func() error) chan error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()
	return ch
}

// Request a given URL and return an io.Reader
func Download(url string, stderr io.Writer) (*http.Response, error) {
	var resp *http.Response
	var err error
	if resp, err = http.Get(url); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New("Got HTTP status code >= 400: " + resp.Status)
	}
	return resp, nil
}

// Debug function, if the debug flag is set, then display. Do nothing otherwise
// If Docker is in damon mode, also send the debug info on the socket
func Debugf(format string, a ...interface{}) {
	if os.Getenv("DEBUG") != "" {

		// Retrieve the stack infos
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "<unknown>"
			line = -1
		} else {
			file = file[strings.LastIndex(file, "/")+1:]
		}

		fmt.Fprintf(os.Stderr, fmt.Sprintf("[debug] %s:%d %s\n", file, line, format), a...)
	}
}

// Reader with progress bar
type progressReader struct {
	reader       io.ReadCloser // Stream to read from
	output       io.Writer     // Where to send progress bar to
	readTotal    int           // Expected stream length (bytes)
	readProgress int           // How much has been read so far (bytes)
	lastUpdate   int           // How many bytes read at least update
	template     string        // Template to print. Default "%v/%v (%v)"
	sf           *StreamFormatter
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	read, err := io.ReadCloser(r.reader).Read(p)
	r.readProgress += read

	updateEvery := 1024 * 512 //512kB
	if r.readTotal > 0 {
		// Update progress for every 1% read if 1% < 512kB
		if increment := int(0.01 * float64(r.readTotal)); increment < updateEvery {
			updateEvery = increment
		}
	}
	if r.readProgress-r.lastUpdate > updateEvery || err != nil {
		if r.readTotal > 0 {
			fmt.Fprintf(r.output, r.template, HumanSize(int64(r.readProgress)), HumanSize(int64(r.readTotal)), fmt.Sprintf("%2.0f%%", float64(r.readProgress)/float64(r.readTotal)*100))
		} else {
			fmt.Fprintf(r.output, r.template, r.readProgress, "?", "n/a")
		}
		r.lastUpdate = r.readProgress
	}
	// Send newline when complete
	if err != nil {
		r.output.Write(r.sf.FormatStatus(""))
	}

	return read, err
}
func (r *progressReader) Close() error {
	return io.ReadCloser(r.reader).Close()
}
func ProgressReader(r io.ReadCloser, size int, output io.Writer, template []byte, sf *StreamFormatter) *progressReader {
	tpl := string(template)
	if tpl == "" {
		tpl = string(sf.FormatProgress("", "%v/%v (%v)"))
	}
	return &progressReader{r, NewWriteFlusher(output), size, 0, 0, tpl, sf}
}

// HumanDuration returns a human-readable approximation of a duration
// (eg. "About a minute", "4 hours ago", etc.)
func HumanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < 1 {
		return "Less than a second"
	} else if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if minutes := int(d.Minutes()); minutes == 1 {
		return "About a minute"
	} else if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	} else if hours := int(d.Hours()); hours == 1 {
		return "About an hour"
	} else if hours < 48 {
		return fmt.Sprintf("%d hours", hours)
	} else if hours < 24*7*2 {
		return fmt.Sprintf("%d days", hours/24)
	} else if hours < 24*30*3 {
		return fmt.Sprintf("%d weeks", hours/24/7)
	} else if hours < 24*365*2 {
		return fmt.Sprintf("%d months", hours/24/30)
	}
	return fmt.Sprintf("%f years", d.Hours()/24/365)
}

// HumanSize returns a human-readable approximation of a size
// using SI standard (eg. "44kB", "17MB")
func HumanSize(size int64) string {
	i := 0
	var sizef float64
	sizef = float64(size)
	units := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	for sizef >= 1000.0 {
		sizef = sizef / 1000.0
		i++
	}
	return fmt.Sprintf("%5.4g %s", sizef, units[i])
}

func Trunc(s string, maxlen int) string {
	if len(s) <= maxlen {
		return s
	}
	return s[:maxlen]
}

// Figure out the absolute path of our own binary
func SelfPath() string {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return path
}

type NopWriter struct{}

func (*NopWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error { return nil }

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

type bufReader struct {
	sync.Mutex
	buf    *bytes.Buffer
	reader io.Reader
	err    error
	wait   sync.Cond
}

func NewBufReader(r io.Reader) *bufReader {
	reader := &bufReader{
		buf:    &bytes.Buffer{},
		reader: r,
	}
	reader.wait.L = &reader.Mutex
	go reader.drain()
	return reader
}

func (r *bufReader) drain() {
	buf := make([]byte, 1024)
	for {
		n, err := r.reader.Read(buf)
		r.Lock()
		if err != nil {
			r.err = err
		} else {
			r.buf.Write(buf[0:n])
		}
		r.wait.Signal()
		r.Unlock()
		if err != nil {
			break
		}
	}
}

func (r *bufReader) Read(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	for {
		n, err = r.buf.Read(p)
		if n > 0 {
			return n, err
		}
		if r.err != nil {
			return 0, r.err
		}
		r.wait.Wait()
	}
}

func (r *bufReader) Close() error {
	closer, ok := r.reader.(io.ReadCloser)
	if !ok {
		return nil
	}
	return closer.Close()
}

type WriteBroadcaster struct {
	sync.Mutex
	writers map[io.WriteCloser]struct{}
}

func (w *WriteBroadcaster) AddWriter(writer io.WriteCloser) {
	w.Lock()
	w.writers[writer] = struct{}{}
	w.Unlock()
}

// FIXME: Is that function used?
// FIXME: This relies on the concrete writer type used having equality operator
func (w *WriteBroadcaster) RemoveWriter(writer io.WriteCloser) {
	w.Lock()
	delete(w.writers, writer)
	w.Unlock()
}

func (w *WriteBroadcaster) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	for writer := range w.writers {
		if n, err := writer.Write(p); err != nil || n != len(p) {
			// On error, evict the writer
			delete(w.writers, writer)
		}
	}
	return len(p), nil
}

func (w *WriteBroadcaster) CloseWriters() error {
	w.Lock()
	defer w.Unlock()
	for writer := range w.writers {
		writer.Close()
	}
	w.writers = make(map[io.WriteCloser]struct{})
	return nil
}

func NewWriteBroadcaster() *WriteBroadcaster {
	return &WriteBroadcaster{writers: make(map[io.WriteCloser]struct{})}
}

func GetTotalUsedFds() int {
	if fds, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/fd", os.Getpid())); err != nil {
		Debugf("Error opening /proc/%d/fd: %s", os.Getpid(), err)
	} else {
		return len(fds)
	}
	return -1
}

// Code c/c from io.Copy() modified to handle escape sequence
func CopyEscapable(dst io.Writer, src io.ReadCloser) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// ---- Docker addition
			// char 16 is C-p
			if nr == 1 && buf[0] == 16 {
				nr, er = src.Read(buf)
				// char 17 is C-q
				if nr == 1 && buf[0] == 17 {
					if err := src.Close(); err != nil {
						return 0, err
					}
					return 0, io.EOF
				}
			}
			// ---- End of docker
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

func HashData(src io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, src); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

type KernelVersionInfo struct {
	Kernel int
	Major  int
	Minor  int
	Flavor string
}

func (k *KernelVersionInfo) String() string {
	flavor := ""
	if len(k.Flavor) > 0 {
		flavor = fmt.Sprintf("-%s", k.Flavor)
	}
	return fmt.Sprintf("%d.%d.%d%s", k.Kernel, k.Major, k.Minor, flavor)
}

// Compare two KernelVersionInfo struct.
// Returns -1 if a < b, = if a == b, 1 it a > b
func CompareKernelVersion(a, b *KernelVersionInfo) int {
	if a.Kernel < b.Kernel {
		return -1
	} else if a.Kernel > b.Kernel {
		return 1
	}

	if a.Major < b.Major {
		return -1
	} else if a.Major > b.Major {
		return 1
	}

	if a.Minor < b.Minor {
		return -1
	} else if a.Minor > b.Minor {
		return 1
	}

	return 0
}

func FindCgroupMountpoint(cgroupType string) (string, error) {
	output, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return "", err
	}

	// /proc/mounts has 6 fields per line, one mount per line, e.g.
	// cgroup /sys/fs/cgroup/devices cgroup rw,relatime,devices 0 0
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) == 6 && parts[2] == "cgroup" {
			for _, opt := range strings.Split(parts[3], ",") {
				if opt == cgroupType {
					return parts[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("cgroup mountpoint not found for %s", cgroupType)
}

// FIXME: this is deprecated by CopyWithTar in archive.go
func CopyDirectory(source, dest string) error {
	if output, err := exec.Command("cp", "-ra", source, dest).CombinedOutput(); err != nil {
		return fmt.Errorf("Error copy: %s (%s)", err, output)
	}
	return nil
}

type NopFlusher struct{}

func (f *NopFlusher) Flush() {}

type WriteFlusher struct {
	w       io.Writer
	flusher http.Flusher
}

func (wf *WriteFlusher) Write(b []byte) (n int, err error) {
	n, err = wf.w.Write(b)
	wf.flusher.Flush()
	return n, err
}

func NewWriteFlusher(w io.Writer) *WriteFlusher {
	var flusher http.Flusher
	if f, ok := w.(http.Flusher); ok {
		flusher = f
	} else {
		flusher = &NopFlusher{}
	}
	return &WriteFlusher{w: w, flusher: flusher}
}

type JSONMessage struct {
	Status   string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
	Error    string `json:"error,omitempty"`
}

type StreamFormatter struct {
	json bool
	used bool
}

func NewStreamFormatter(json bool) *StreamFormatter {
	return &StreamFormatter{json, false}
}

func (sf *StreamFormatter) FormatStatus(format string, a ...interface{}) []byte {
	sf.used = true
	str := fmt.Sprintf(format, a...)
	if sf.json {
		b, err := json.Marshal(&JSONMessage{Status: str})
		if err != nil {
			return sf.FormatError(err)
		}
		return b
	}
	return []byte(str + "\r\n")
}

func (sf *StreamFormatter) FormatError(err error) []byte {
	sf.used = true
	if sf.json {
		if b, err := json.Marshal(&JSONMessage{Error: err.Error()}); err == nil {
			return b
		}
		return []byte("{\"error\":\"format error\"}")
	}
	return []byte("Error: " + err.Error() + "\r\n")
}

func (sf *StreamFormatter) FormatProgress(action, str string) []byte {
	sf.used = true
	if sf.json {
		b, err := json.Marshal(&JSONMessage{Status: action, Progress: str})
		if err != nil {
			return nil
		}
		return b
	}
	return []byte(action + " " + str + "\r")
}

func (sf *StreamFormatter) Used() bool {
	return sf.used
}

func IsURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

func IsGIT(str string) bool {
	return strings.HasPrefix(str, "git://") || strings.HasPrefix(str, "github.com/")
}

func CheckLocalDns() bool {
	resolv, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		Debugf("Error openning resolv.conf: %s", err)
		return false
	}
	for _, ip := range []string{
		"127.0.0.1",
		"127.0.1.1",
	} {
		if strings.Contains(string(resolv), ip) {
			return true
		}
	}
	return false
}

func ParseHost(host string, port int, addr string) string {
	if strings.HasPrefix(addr, "unix://") {
		return addr
	}
	if strings.HasPrefix(addr, "tcp://") {
		addr = strings.TrimPrefix(addr, "tcp://")
	}
	if strings.Contains(addr, ":") {
		hostParts := strings.Split(addr, ":")
		if len(hostParts) != 2 {
			log.Fatal("Invalid bind address format.")
			os.Exit(-1)
		}
		if hostParts[0] != "" {
			host = hostParts[0]
		}
		if p, err := strconv.Atoi(hostParts[1]); err == nil {
			port = p
		}
	} else {
		host = addr
	}
	return fmt.Sprintf("tcp://%s:%d", host, port)
}
