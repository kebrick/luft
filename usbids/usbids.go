package usbids

import (
	"bufio"
	"github.com/i582/cfmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	vendors = map[string]*Vendor{}

	version    = regexp.MustCompile(`Version: (\d{4}.\d{2}.\d{2})`)
	date       = regexp.MustCompile(`Date:\s+(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)
	vendorLine = regexp.MustCompile(`^([[:xdigit:]]{4})\s{2}(.+)$`)
	deviceLine = regexp.MustCompile(`\t([[:xdigit:]]{4})\s{2}(.+)$`)

	Ids     = []string{"/var/lib/usbutils/usb.ids", "/usr/share/hwdata/usb.ids", "usb.ids"}
	Version = ""
	Date    = ""
)

type Vendor struct {
	ID     string
	Name   string
	Device map[string]*Device
}

type Device struct {
	ID   string
	Name string
}

func LoadFromFiles() error {
	for _, usbID := range Ids {
		if err := LoadFromFile(usbID); err != nil {
			continue
		}
		return nil
	}
	return nil
}

func ParseUsbIDs(file *os.File) error {

	scanner := bufio.NewScanner(file)

	emitVendor := func(vendors map[string]*Vendor, vendor Vendor) {
		vendors[vendor.ID] = &vendor
	}

	var currVendor *Vendor
	var prevVendor *Vendor

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, `#`) {
			if result := version.FindStringSubmatch(line); len(result) != 0 {
				Version = result[1]
			}
			if result := date.FindStringSubmatch(line); len(result) != 0 {
				Date = result[1]
			}
			continue
		} else if result := vendorLine.FindStringSubmatch(line); len(result) != 0 {
			if vendor := prevVendor; vendor != nil {
				emitVendor(vendors, *vendor)
			}
			currVendor = &Vendor{
				Name:   result[2],
				ID:     result[1],
				Device: map[string]*Device{},
			}
			prevVendor = currVendor
		} else if result := deviceLine.FindStringSubmatch(line); len(result) != 0 {
			if currVendor := currVendor; currVendor != nil {
				currVendor.Device[result[1]] = &Device{
					ID:   result[1],
					Name: result[2],
				}
			}
		} else {
			break
		}
	}

	if scanner.Err() != nil {
		_, _ = cfmt.Println(cfmt.Sprintf("{{Error while parse usb.ids}}::red"))
	}

	_, _ = cfmt.Println(cfmt.Sprintf("{{[%v] usb.ids loaded from: %s, Version: %s, Date: %s}}::green", time.Now().Format(time.Stamp), file.Name(), Version, Date))
	_, _ = cfmt.Println(cfmt.Sprintf("{{[%v] usb.ids %d vendors load}}::green", time.Now().Format(time.Stamp), len(vendors)))
	return nil
}

func LoadFromFile(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	return ParseUsbIDs(file)
}

func FindDevice(vid, pid string) (string, string) {
	if vendors := vendors; vendors != nil {
		vendor := vendors[vid]
		if vendor != nil {
			device := vendor.Device[pid]
			if device != nil {
				return vendor.Name, device.Name
			}
			return vendor.Name, ""
		}
		return "", ""
	}
	return "", ""
}
