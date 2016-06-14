package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"upnp"

	"github.com/apcera/termtables"
)

func clearConsole() {
	//print("\033[H\033[2J")
}
func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	str, _ := reader.ReadString('\n')
	return strings.TrimSpace(str)
}
func getInterface() (*net.Interface, error) {
	addrs, _ := net.Interfaces()
	for n, addr := range addrs {
		ip, err := addr.Addrs()
		if err != nil {
			continue
		}
		fmt.Println(n, " : ", addr.Name, " {", addr.HardwareAddr, "} ", ip)
	}
	fmt.Print("Please select interface number : ")
	interfaceNum, err := strconv.ParseUint(getInput(), 10, 64)
	if err != nil {
		return nil, err
	}
	return &addrs[interfaceNum], nil
}
func main() {
	//pcp.LaunchPCP()
	clearConsole()
	addrs, _ := net.Interfaces()
	for n, addr := range addrs {
		ip, err := addr.Addrs()
		if err != nil {
			continue
		}
		fmt.Println(n, " : ", addr.Name, " {", addr.HardwareAddr, "} ", ip)
	}
	fmt.Print("Please select interface number : ")
	interfaceNum, err := strconv.ParseUint(getInput(), 10, 64)
	if err != nil {
		log.Println(err)
	}
	for {
		clearConsole()
		fmt.Println("1 : Upnp tools")
		fmt.Println("2 : Show ipv4")
		fmt.Println("3 : Show WAN status")
		fmt.Println("4 : Show nat")
		fmt.Println("5 : Add nat")
		fmt.Println("q : Quit")
		fmt.Print("Please select menu :")
		input := getInput()
		if input == "q" {
			return
		}
		menuSelect, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			log.Println(err)
		}
		switch menuSelect {
		case 1:
			upnpTools(addrs[interfaceNum])
			break
		case 2:
			ipv4(addrs[interfaceNum])
			break
		case 3:
			wanStatus(addrs[interfaceNum])
			break
		case 4:
			showNat(addrs[interfaceNum])
			break
		case 5:
			addNat(addrs[interfaceNum])
		}
	}
}
func wanStatus(_interface net.Interface) {
	clearConsole()
	up := upnp.NewUPNP(upnp.SERVICE_GATEWAY_STATE)
	devices := up.GetAllCompatibleDevice(_interface, 1)
	if len(devices) == 0 {
		fmt.Println("not found service")
		return
	}
	services := devices[0].GetServicesByType(upnp.SERVICE_GATEWAY_STATE)
	if len(services) == 0 {
		fmt.Println("not found service")
		return
	}
	action := services[0].GetAction("GetCommonLinkProperties")
	resp, err := action.Send()
	if err != nil {
		log.Println(err)
	}
	WANAccessType, _ := resp.GetValueArgument("NewWANAccessType")
	PhysicalLinkStatus, _ := resp.GetValueArgument("NewPhysicalLinkStatus")

	fmt.Println("WANAccessType : ", WANAccessType)
	fmt.Println("PhysicalLinkStatus : ", PhysicalLinkStatus)
	fmt.Print("Press enter")
	getInput()
}
func showNat(_interface net.Interface) {
	clearConsole()
	service := ipv4Gateway(_interface, upnp.SERVICE_GATEWAY_IPV4_V2)
	if service == nil {
		service = ipv4Gateway(_interface, upnp.SERVICE_GATEWAY_IPV4_V1)
	}
	action := service.GetAction("GetGenericPortMappingEntry")
	if action == nil {
		return
	}
	table := termtables.CreateTable()
	table.AddHeaders("Enabled", "Protocol", "Internal Client", "Internal Port", "External Port", "Time expired", "Description")
	//fmt.Println("enabled | protocol | internal Port | internalClient | external Port | time | description")
	for i := 0; ; i++ {
		action.AddVariable("NewPortMappingIndex", strconv.FormatInt(int64(i), 10))
		rep, err := action.Send()
		if err != nil {
			fmt.Println(err)
			break
		}
		if !rep.Success() {
			if rep.GetError().ErrorCode == "713" {
				break
			}
			fmt.Println(rep.GetError().ErrorCode)
			fmt.Println(rep.GetError().ErrorDescription)
			break
		}
		protocol, _ := rep.GetValueArgument("NewProtocol")
		internalPort, _ := rep.GetValueArgument("NewInternalPort")
		internalClient, _ := rep.GetValueArgument("NewInternalClient")
		externalPort, _ := rep.GetValueArgument("NewExternalPort")
		enabled, _ := rep.GetValueArgument("NewEnabled")
		portMappingDescription, _ := rep.GetValueArgument("NewPortMappingDescription")
		leaseDuration, _ := rep.GetValueArgument("NewLeaseDuration")
		table.AddRow(enabled, protocol, internalClient, internalPort, externalPort, leaseDuration, portMappingDescription)
	}
	fmt.Println(table.Render())
	fmt.Print("Press enter")
	getInput()
}
func addNat(_interface net.Interface) {
	fmt.Println("Please complete this :")
	fmt.Print("External port : ")
	extPort := getInput()
	fmt.Print("Internal port default(", extPort, "): ")
	intPort := getInput()
	if intPort == "" {
		intPort = extPort
	}
	ip := upnp.GetIPAdress(_interface)
	fmt.Print("Internal Network ip default(", ip, "): ")
	intIp := getInput()
	if intIp == "" {
		intIp = ip
	}
	var protocol string
	for {
		fmt.Print("Set protocol TCP or UDP: ")
		protocol = getInput()
		if protocol == "TCP" || protocol == "UDP" {
			break
		} else {
			fmt.Println("You value is not valide : ", protocol)
		}
	}
	fmt.Print("Timeout default(0): ")
	timeout := getInput()
	if timeout == "" {
		timeout = "0"
	}
	fmt.Print("Description default(upnptools): ")
	description := getInput()
	if description == "" {
		description = "upnptools"
	}
	service := ipv4Gateway(_interface, upnp.SERVICE_GATEWAY_IPV4_V2)
	if service == nil {
		service = ipv4Gateway(_interface, upnp.SERVICE_GATEWAY_IPV4_V1)
	}
	action := service.GetAction("AddPortMapping")
	if action == nil {
		return
	}
	action.AddVariable("NewRemoteHost", "")
	action.AddVariable("NewExternalPort", extPort)
	action.AddVariable("NewProtocol", protocol)
	action.AddVariable("NewInternalPort", intPort)
	action.AddVariable("NewInternalClient", intIp)
	action.AddVariable("NewEnabled", "1")
	action.AddVariable("NewPortMappingDescription", description)
	action.AddVariable("NewLeaseDuration", timeout)
	resp, err := action.Send()
	if !resp.Success() || err != nil {
		fmt.Print("Error port is not mapping : ")
		if err != nil {
			fmt.Print(err)
		} else {
			fmt.Print(resp.GetError().ErrorDescription)
		}
	} else {
		fmt.Print("Port is mapping")
	}
	getInput()
}
func ipv4Gateway(_interface net.Interface, typeService string) *upnp.Service {
	up := upnp.NewUPNP(typeService)
	devices := up.GetAllCompatibleDevice(_interface, 1)
	if len(devices) == 0 {
		return nil
	}
	services := devices[0].GetServicesByType(typeService)
	if len(services) == 0 {
		return nil
	}

	return services[0]

}

//Exemple for get external ipv4 address from gateway
func ExempleNewUPNP() {
	up := upnp.NewUPNP(upnp.SERVICE_GATEWAY_IPV4_V2)
	_interface, err := getInterface()
	if err != nil {
		panic(err)
	}
	devices := up.GetAllCompatibleDevice(*_interface, 1)
	if len(devices) == 0 {
		return
	}
	services := devices[0].GetServicesByType(upnp.SERVICE_GATEWAY_IPV4_V2)
	if len(services) == 0 {
		return
	}
	service := services[0]
	response, err := service.GetAction("GetExternalIPAddress").Send()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(response.ToString())
	fmt.Print("Press enter")
	getInput()

}
func ipv4(_interface net.Interface) {
	clearConsole()
	service := ipv4Gateway(_interface, upnp.SERVICE_GATEWAY_IPV4_V2)
	if service == nil {
		service = ipv4Gateway(_interface, upnp.SERVICE_GATEWAY_IPV4_V1)

	}
	scpd, _ := service.GetSCPD()
	fmt.Println(scpd.URL, "--", service.SCPDURL)
	if service == nil {
		fmt.Println("not found service")
		return
	}

	response, err := service.GetAction("GetExternalIPAddress").Send()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(response.ToString())
	fmt.Print("Press enter")
	getInput()
}

func upnpTools(_interface net.Interface) {
	clearConsole()
	fmt.Println("Scanning...")
	up := upnp.NewUPNPAllService()
	devices := up.GetAllCompatibleDevice(_interface, 5)
	for {
		clearConsole()
		for n, device := range devices {
			fmt.Println(n, " : ", device.FriendlyName, " {", device.SerialNumber, "}")
		}
		fmt.Println("r : retour")
		fmt.Print("Please select device number : ")
		str := getInput()
		if str == "r" {
			return
		}
		deviceNumber, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			log.Println(err)
		}
		clearConsole()
		launchUpnp(devices[int(deviceNumber)])
	}

}
func selectService(device *upnp.Device) {
	for {
		clearConsole()
		services := device.GetAllService()
		for n, service := range services {
			fmt.Println(n, " : ", service.ServiceType)
		}
		fmt.Println("r : retour")
		fmt.Print("Please select service number : ")
		input := getInput()
		if input == "r" {
			return
		}
		serviceNumber, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			log.Println(err)
		}
		clearConsole()
		service := services[serviceNumber]
		if err != nil {
			log.Println(err)
			return
		}
		SetAction(service)
	}
}
func SetAction(service *upnp.Service) {
	for {
		actions := service.GetActions()
		for n, action := range actions {
			fmt.Println(n, " : ", action.GetName())
		}
		fmt.Println("r : retour")
		fmt.Print("Please select action number : ")
		input := getInput()
		if input == "r" {
			return
		}
		actionNumber, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		clearConsole()
		action := actions[actionNumber]
		args := action.GetInArguments()
		for _, argument := range args {
			if argument.GetType() != "" {
				atmp := argument.GetAllowedValues()

				if len(atmp) > 0 {
					fmt.Print(argument.GetName(), " (", strings.Join(atmp, ","), ") [", argument.GetDefault(), "]= ")
				} else {
					fmt.Print(argument.GetName(), " (", argument.GetType(), ") [", argument.GetDefault(), "]= ")
				}
			} else {
				fmt.Print(argument.GetName(), "= ")
			}
			action.AddVariable(argument.GetName(), getInput())
		}
		rep, err := action.Send()
		if err != nil {
			log.Println(err)
			return
		}
		if rep.Success() {
			fmt.Println(rep.ToString())
		} else {
			errorUpnp := rep.GetError()
			fmt.Println("Code : ", errorUpnp.ErrorCode)
			fmt.Println("Description : ", errorUpnp.ErrorDescription)
		}
		getInput()
		clearConsole()
	}
}
func launchUpnp(device *upnp.Device) {
	clearConsole()
	selectService(device)

}
