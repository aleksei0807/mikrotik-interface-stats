package main

import (
	"net"

	"github.com/gramework/gramework"
	"github.com/soniah/gosnmp"
)

const targetIP = "10.0.0.1"

var oids = map[string]string{
	"name":         ".1.3.6.1.2.1.2.2.1.2.",
	"actual-mtu":   ".1.3.6.1.2.1.2.2.1.4.",
	"mac-address":  ".1.3.6.1.2.1.2.2.1.6.",
	"admin-status": ".1.3.6.1.2.1.2.2.1.7.",
	"oper-status":  ".1.3.6.1.2.1.2.2.1.8.",
	"bytes-in":     ".1.3.6.1.2.1.31.1.1.1.6.",
	"packets-in":   ".1.3.6.1.2.1.31.1.1.1.7.",
	"discards-in":  ".1.3.6.1.2.1.2.2.1.13.",
	"errors-in":    ".1.3.6.1.2.1.2.2.1.14.",
	"bytes-out":    ".1.3.6.1.2.1.31.1.1.1.10.",
	"packets-out":  ".1.3.6.1.2.1.31.1.1.1.11.",
	"discards-out": ".1.3.6.1.2.1.2.2.1.19.",
	"errors-out":   ".1.3.6.1.2.1.2.2.1.20.",
}

var numbers = []string{"1", "2", "3", "4", "5", "6", "11", "12", "7", "9", "13"}

func getData() (map[string]map[string]string, error) {
	var result = make(map[string]map[string]string)

	gosnmp.Default.Target = targetIP

	if err := gosnmp.Default.Connect(); err != nil {
		return nil, err
	}

	defer gosnmp.Default.Conn.Close()

	for key, value := range oids {
		localOIDs := []string{}
		for _, number := range numbers {
			localOIDs = append(localOIDs, value+number)
		}

		res, err := gosnmp.Default.Get(localOIDs)
		if err != nil {
			return nil, err
		}

		for i, variable := range res.Variables {
			if result[numbers[i]] == nil {
				result[numbers[i]] = make(map[string]string)
			}

			// the Value of each variable returned by Get() implements
			// interface{}. You could do a type switch...
			switch variable.Type {
			case gosnmp.OctetString:
				bytes := variable.Value.([]byte)
				if key == "mac-address" && len(bytes) == 6 {
					result[numbers[i]][key] = net.HardwareAddr(bytes).String()
					continue
				}
				result[numbers[i]][key] = string(bytes)
			default:
				// ... or often you're just interested in numeric values.
				// ToBigInt() will return the Value as a BigInt, for plugging
				// into your calculations.
				result[numbers[i]][key] = gosnmp.ToBigInt(variable.Value).String()
			}
		}
	}

	return result, nil
}

func main() {
	app := gramework.New()
	app.GET("/", func(ctx *gramework.Context) (interface{}, error) {
		data, err := getData()
		if err != nil {
			gramework.Logger.WithError(err).Error("getData() error")
		} else {
			gramework.Logger.Infof("%#v", data)
		}
		return data, err
	})
	app.ListenAndServe()
}
