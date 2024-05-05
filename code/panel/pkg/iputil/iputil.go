package iputil

import (
	"errors"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type IpRange struct {
	Start string
	End   string
}

// IsValidIP checks if the given IP address is valid
func IsValidIP(w http.ResponseWriter, inputIP string) bool {
	if ip := net.ParseIP(inputIP); ip != nil {
		return true
	}

	return false
}

// IsValidIPRange checks if the given IP range is valid
func IsValidIPRange(w http.ResponseWriter, ipRange string) bool {

	ipList, err := ConvertIPRangeToIPSize(w, ipRange)
	if err != nil {
		return false
	}
	if ipList != nil && ipList.Int64() > 0 {
		return true
	}

	return false

}

func ConvertIPRangeToIPSize(w http.ResponseWriter, ipRange string) (*big.Int, error) {
	if ipRange == "" {
		return nil, nil
	}

	ipRange = strings.TrimSpace(ipRange)

	if strings.Contains(ipRange, "-") {
		return nil, errors.New("Invalid IP range format")
	} else if strings.Contains(ipRange, "/") {
		ipRangeSplitted := strings.Split(ipRange, "/")
		cidrPrefix := ipRangeSplitted[0]
		_, err := strconv.Atoi(ipRangeSplitted[1])
		if err != nil {
			return nil, nil
		}

		if ip := net.ParseIP(cidrPrefix); ip == nil {
			return nil, errors.New("Invalid IP address")
		}

		return CIDRRangeSize(ipRange), nil
	}

	return nil, nil
}

// CIDRRangeSize calculates the total number of IP addresses in a CIDR range.
func CIDRRangeSize(cidr string) *big.Int {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	// Get the number of bits in the mask
	ones, bits := ipNet.Mask.Size()

	// Calculate the number of possible IPs
	// The formula is 2^(bits - ones)
	totalIPs := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(bits-ones)), nil)

	// Subtract 2 for network and broadcast addresses
	totalIPs.Sub(totalIPs, big.NewInt(2))

	return totalIPs
}
