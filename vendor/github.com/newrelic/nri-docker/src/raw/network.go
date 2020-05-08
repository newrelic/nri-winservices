package raw

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/log"
)

// NetworkFetcher fetches the network metrics from the /proc file system
// TODO: use cgroups library
type networkFetcher struct {
	hostRoot string
}

func newNetworkFetcher(hostRoot string) *networkFetcher {
	return &networkFetcher{hostRoot: hostRoot}
}

func (f *networkFetcher) Fetch(containerPid int) (Network, error) {
	var network Network
	filePath := containerToHost(f.hostRoot, path.Join("/proc", strconv.Itoa(containerPid), "net", "dev"))
	file, err := os.Open(filePath)
	if err != nil {
		return network, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	sc.Split(bufio.ScanLines)
	sc.Scan() // scan first header line
	sc.Scan() // scan second header line
	for sc.Scan() {
		ws := bufio.NewScanner(strings.NewReader(sc.Text()))
		ws.Split(bufio.ScanWords)
		words := make([]string, 0, 18)
		for ws.Scan() {
			words = append(words, ws.Text())
		}
		if len(words) < 13 {
			log.Debug("apparently malformed line: %s", sc.Text())
			continue
		}
		if strings.HasPrefix(words[0], "lo") { // ignoring loopback
			continue
		}

		rxBytes, err := strconv.Atoi(words[1])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		rxPackets, err := strconv.Atoi(words[2])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		rxErrors, err := strconv.Atoi(words[3])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		rxDropped, err := strconv.Atoi(words[4])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		txBytes, err := strconv.Atoi(words[9])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		txPackets, err := strconv.Atoi(words[10])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		txErrors, err := strconv.Atoi(words[11])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}
		txDropped, err := strconv.Atoi(words[12])
		if err != nil {
			log.Debug("apparently malformed line %q. Cause: %s", sc.Text(), err.Error())
			continue
		}

		network.RxBytes += int64(rxBytes)
		network.RxDropped += int64(rxDropped)
		network.RxErrors += int64(rxErrors)
		network.RxPackets += int64(rxPackets)
		network.TxBytes += int64(txBytes)
		network.TxDropped += int64(txDropped)
		network.TxErrors += int64(txErrors)
		network.TxPackets += int64(txPackets)
	}

	return network, nil
}
