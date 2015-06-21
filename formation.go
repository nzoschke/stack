package formation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
)

type ManifestEntry struct {
	Command interface{} `yaml:"command"`
	Links   []string    `yaml:"links"`
	Ports   []string    `yaml:"ports"`
	Volumes []string    `yaml:"volumes"`

	Randoms []string
}

type Manifest map[string]ManifestEntry

type Listener struct {
	Balancer string
	Process  string
}

func buildTemplate(data interface{}) (string, error) {
	tmpl, err := template.New("formation").Funcs(templateHelpers()).ParseFiles("staging.tmpl")

	if err != nil {
		return "", err
	}

	var formation bytes.Buffer

	err = tmpl.Execute(&formation, data)

	if err != nil {
		return "", err
	}

	return formation.String(), nil
}

func templateHelpers() template.FuncMap {
	return template.FuncMap{
		"command": func(command interface{}) string {
			switch cmd := command.(type) {
			case nil:
				return ""
			case string:
				return cmd
			case []interface{}:
				parts := make([]string, len(cmd))

				for i, c := range cmd {
					parts[i] = c.(string)
				}

				return strings.Join(parts, " ")
			default:
				fmt.Fprintf(os.Stderr, "unexpected type for command: %T\n", cmd)
			}
			return ""
		},
		"ingress": func(m Manifest) template.HTML {
			ls := []string{}

			for ps, entry := range m {
				for _, port := range entry.Ports {
					parts := strings.SplitN(port, ":", 2)

					if len(parts) != 2 {
						continue
					}

					ls = append(ls, fmt.Sprintf(`{ "CidrIp": "0.0.0.0/0", "IpProtocol": "tcp", "FromPort": { "Ref": "%sPort%sBalancer" }, "ToPort": { "Ref": "%sPort%sBalancer" } }`, upperName(ps), parts[0], upperName(ps), parts[0]))
				}
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"links": func(m Manifest) template.HTML {
			links := []string{}

			for ps, _ := range m {
				links = append(links, fmt.Sprintf(`{ "Fn::If": [ "Blank%sService", "%s:%s", { "Ref": "AWS::NoValue" } ] }`, upperName(ps), ps, ps))
			}

			return template.HTML(strings.Join(links, ","))
		},
		"listeners": func(m Manifest) template.HTML {
			ls := []string{}

			for ps, entry := range m {
				for _, port := range entry.Ports {
					parts := strings.SplitN(port, ":", 2)

					if len(parts) != 2 {
						continue
					}

					ls = append(ls, fmt.Sprintf(`{ "Protocol": "TCP", "LoadBalancerPort": { "Ref": "%sPort%sBalancer" }, "InstanceProtocol": "TCP", "InstancePort": { "Ref": "%sPort%sHost" } }`, upperName(ps), parts[0], upperName(ps), parts[0]))
				}
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"loadbalancers": func(m Manifest) template.HTML {
			ls := []string{}

			for ps, entry := range m {
				for _, port := range entry.Ports {
					parts := strings.SplitN(port, ":", 2)

					if len(parts) != 2 {
						continue
					}

					ls = append(ls, fmt.Sprintf(`{ "Fn::Join": [ ":", [ { "Ref": "Balancer" }, "%s", "%s" ] ] }`, ps, parts[1]))
				}
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"names": func(m Manifest) template.HTML {
			names := []string{}

			for ps, _ := range m {
				names = append(names, fmt.Sprintf(`{ "Fn::If": [ "Blank%sService", "%s", { "Ref": "AWS::NoValue" } ] }`, upperName(ps), ps))
			}

			return template.HTML(strings.Join(names, ","))
		},
		"split": func(ss string, t string) []string {
			return strings.Split(ss, t)
		},
		"tasks": func(m Manifest) template.HTML {
			ls := []string{}

			for ps, entry := range m {
				mappings := []string{}

				for _, port := range entry.Ports {
					parts := strings.SplitN(port, ":", 2)

					if len(parts) != 2 {
						continue
					}

					mappings = append(mappings, fmt.Sprintf(`{ "Fn::Join": [ ":", [ { "Ref": "%sPort%sHost" }, "%s" ] ] }`, upperName(ps), parts[0], parts[1]))
				}

				links := make([]string, len(entry.Links))

				for i, link := range entry.Links {
					links[i] = fmt.Sprintf(`{ "Fn::If": [ "Blank%sService",
            "%s:%s",
            { "Ref" : "AWS::NoValue" } ] }`, upperName(link), link, link)
				}

				services := make([]string, len(entry.Links))

				for i, link := range entry.Links {
					services[i] = fmt.Sprintf(`{ "Fn::If": [ "Blank%sService",
            { "Ref" : "AWS::NoValue" },
            { "Fn::Join": [ ":", [ { "Ref" : "%sService" }, "%s" ] ] } ] }`, upperName(link), upperName(link), link)
				}

				volumes := []string{}

				for _, volume := range entry.Volumes {
					if strings.HasPrefix(volume, "/var/run/docker.sock") {
						volumes = append(volumes, fmt.Sprintf(`"%s"`, volume))
					}
				}

				ls = append(ls, fmt.Sprintf(`{ "Fn::If": [ "Blank%sService",
        {
          "Name": "%s",
          "Image": { "Ref": "%sImage" },
          "Command": { "Ref": "%sCommand" },
          "Key": { "Ref": "Key" },
          "CPU": "200",
          "Memory": "300",
          "Links": [ %s ],
          "Volumes": [ %s ],
          "Services": [ %s ],
          "PortMappings": [ %s ]
        }, { "Ref" : "AWS::NoValue" } ] }`, upperName(ps), ps, upperName(ps), upperName(ps), strings.Join(links, ","), strings.Join(volumes, ","), strings.Join(services, ","), strings.Join(mappings, ",")))
			}

			return template.HTML(strings.Join(ls, ","))
		},
		"upper": func(name string) string {
			return upperName(name)
		},
	}
}

func upperName(name string) string {
	us := strings.ToUpper(name[0:1]) + name[1:]

	for {
		i := strings.Index(us, "-")

		if i == -1 {
			break
		}

		s := us[0:i]

		if len(us) > i+1 {
			s += strings.ToUpper(us[i+1 : i+2])
		}

		if len(us) > i+2 {
			s += us[i+2:]
		}

		us = s
	}

	return us
}

// base types
type Hash map[string]interface{}
type List []interface{}

// special types

type Template struct {
	AWSTemplateFormatVersion string
	Description              string
	Resources                Resources
}

type InternetGateway struct {
}

func (g InternetGateway) Type() string {
	return "AWS::EC2::InternetGateway"
}

type Vpc struct {
	CidrBlock        interface{} `json:"CidrBlock,omitempty"`
	EnableDnsSupport interface{} `json:"EnableDnsSupport,omitempty"`
	InstanceTenancy  interface{} `json:"InstanceTenancy,omitempty"`
	Tags             interface{} `json:"Tags,omitempty"`
}

type VPCGatewayAttachment struct {
	InternetGatewayId interface{} `json:"InternetGatewayId,omitempty"`
	VpcId             interface{} `json:"VpcId,omitempty"`
}

type RouteTable struct {
	VpcId interface{} `json:"VpcId,omitempty"`
}

type SecurityGroup struct {
	GroupDescription     interface{} `json:"GroupDescription,omitempty"`
	SecurityGroupEgress  interface{} `json:"SecurityGroupEgress,omitempty"`
	SecurityGroupIngress interface{} `json:"SecurityGroupIngress,omitempty"`
	VpcId                interface{} `json:"VpcId,omitempty"`
	Tags                 interface{} `json:"Tags,omitempty"`
}

func (s SecurityGroup) Type() string {
	return "AWS::EC2::SecurityGroup"
}

type SecurityGroupEgress struct {
	CidrIp                     interface{} `json:"CidrIp,omitempty"`
	FromPort                   interface{} `json:"FromPort,omitempty"`
	IpProtocol                 interface{} `json:"IpProtocol,omitempty"`
	DestinationSecurityGroupId interface{} `json:"DestinationSecurityGroupId,omitempty"`
	ToPort                     interface{} `json:"ToPort,omitempty"`
}

type SecurityGroupIngress struct {
	CidrIp                     interface{} `json:"CidrIp,omitempty"`
	FromPort                   interface{} `json:"FromPort,omitempty"`
	IpProtocol                 interface{} `json:"IpProtocol,omitempty"`
	SourceSecurityGroupId      interface{} `json:"SourceSecurityGroupId,omitempty"`
	SourceSecurityGroupName    interface{} `json:"SourceSecurityGroupName,omitempty"`
	SourceSecurityGroupOwnerId interface{} `json:"SourceSecurityGroupOwnerId,omitempty"`
	ToPort                     interface{} `json:"ToPort,omitempty"`
}

func (r RouteTable) Type() string {
	return "AWS::EC2::RouteTable"
}

func (e VPCGatewayAttachment) Type() string {
	return "AWS::EC2::VPCGatewayAttachment"
}

type Route struct {
	DestinationCidrBlock interface{} `json:"DestinationCidrBlock,omitempty"`
	GatewayId            interface{} `json:"GatewayId,omitempty"`
	RouteTableId         interface{} `json:"RouteTableId,omitempty"`
}

func (r Route) Type() string {
	return "AWS::EC2::Route"
}

type Subnet struct {
	AvailabilityZone interface{} `json:"AvailabilityZone,omitempty"`
	CidrBlock        interface{} `json:"CidrBlock,omitempty"`
	VpcId            interface{} `json:"VpcId,omitempty"`
}

type SubnetRouteTableAssociation struct {
	RouteTableId interface{} `json:"RouteTableId,omitempty"`
	SubnetId     interface{} `json:"SubnetId,omitempty"`
}

func (s SubnetRouteTableAssociation) Type() string {
	return "AWS::EC2::SubnetRouteTableAssociation"
}

func (s Subnet) Type() string {
	return "AWS::EC2::Subnet"
}

func (r Resources) MarshalJSON() ([]byte, error) {
	lines := []string{}
	for k, v := range r {
		kj, e := json.Marshal(k)
		if e != nil {
			return nil, e
		}
		p := map[string]interface{}{
			"Type": v,
		}
		vj, e := json.Marshal(v)
		if e != nil {
			return nil, e
		}
		if string(vj) != "{}" {
			p["Properties"] = v
		}
		pj, e := json.Marshal(p)
		lines = append(lines, string(kj)+": "+string(pj))
	}
	return []byte("{" + strings.Join(lines, ",\n") + "}"), nil
}

func (v Vpc) Type() string {
	return "AWS::EC2::VPC"
}

type Instance struct {
	AvailabilityZone      interface{} `json:"AvailabilityZone,omitempty"`
	BlockDeviceMappings   interface{} `json:"BlockDeviceMappings,omitempty"`
	DisableApiTermination interface{} `json:"DisableApiTermination,omitempty"`
	EbsOptimized          interface{} `json:"EbsOptimized,omitempty"`
	IamInstanceProfile    interface{} `json:"IamInstanceProfile,omitempty"`
	ImageId               interface{} `json:"ImageId,omitempty"`
	InstanceType          interface{} `json:"InstanceType,omitempty"`
	KernelId              interface{} `json:"KernelId,omitempty"`
	KeyName               interface{} `json:"KeyName,omitempty"`
	Monitoring            interface{} `json:"Monitoring,omitempty"`
	NetworkInterfaces     interface{} `json:"NetworkInterfaces,omitempty"`
	PlacementGroupName    interface{} `json:"PlacementGroupName,omitempty"`
	PrivateIpAddress      interface{} `json:"PrivateIpAddress,omitempty"`
	RamdiskId             interface{} `json:"RamdiskId,omitempty"`
	SecurityGroupIds      interface{} `json:"SecurityGroupIds,omitempty"`
	SecurityGroups        interface{} `json:"SecurityGroups,omitempty"`
	SourceDestCheck       interface{} `json:"SourceDestCheck,omitempty"`
	SubnetId              interface{} `json:"SubnetId,omitempty"`
	Tags                  interface{} `json:"Tags,omitempty"`
	Tenancy               interface{} `json:"Tenancy,omitempty"`
	UserData              interface{} `json:"UserData,omitempty"`
	Volumes               interface{} `json:"Volumes,omitempty"`
}

func (i Instance) Type() string {
	return "AWS::EC2::Instance"
}

type NetworkInterface struct {
	AssociatePublicIpAddress       interface{} `json:"AssociatePublicIpAddress,omitempty"`
	DeleteOnTermination            interface{} `json:"DeleteOnTermination,omitempty"`
	Description                    interface{} `json:"Description,omitempty"`
	DeviceIndex                    interface{} `json:"DeviceIndex,omitempty"`
	GroupSet                       interface{} `json:"GroupSet,omitempty"`
	NetworkInterfaceId             interface{} `json:"NetworkInterfaceId,omitempty"`
	PrivateIpAddress               interface{} `json:"PrivateIpAddress,omitempty"`
	PrivateIpAddresses             interface{} `json:"PrivateIpAddresses,omitempty"`
	SecondaryPrivateIpAddressCount interface{} `json:"SecondaryPrivateIpAddressCount,omitempty"`
	SubnetId                       interface{} `json:"SubnetId,omitempty"`
}

type Tag struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

type Resources map[string]ResourceType

type ResourceType map[string]interface{}

func ref(i interface{}) Hash {
	return Hash{"Ref": i}
}
