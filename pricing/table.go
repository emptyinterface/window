package pricing

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	Table struct {
		Rows []*Row
	}

	Row struct {
		OfferCode string
		Region    string

		Availability                string    `json:",omitempty"`
		CacheEngine                 string    `json:",omitempty"`
		ClockSpeed                  float64   `json:",omitempty"`
		ContentType                 string    `json:",omitempty"`
		Currency                    string    `json:",omitempty"`
		CurrentGeneration           bool      `json:",omitempty"`
		DatabaseEdition             string    `json:",omitempty"`
		DatabaseEngine              string    `json:",omitempty"`
		DedicatedEBSThroughput      string    `json:",omitempty"`
		DeploymentOption            string    `json:",omitempty"`
		Description                 string    `json:",omitempty"`
		Durability                  string    `json:",omitempty"`
		EBSOptimized                bool      `json:",omitempty"`
		ECU                         int       `json:",omitempty"`
		EffectiveDate               time.Time `json:",omitempty"`
		EndingRange                 float64   `json:",omitempty"`
		EndpointType                string    `json:",omitempty"`
		EngineCode                  string    `json:",omitempty"`
		EnhancedNetworkingSupported bool      `json:",omitempty"`
		FeeCode                     string    `json:",omitempty"`
		FeeDescription              string    `json:",omitempty"`
		FromLocation                string    `json:",omitempty"`
		FromLocationType            string    `json:",omitempty"`
		GPU                         int       `json:",omitempty"`
		Group                       string    `json:",omitempty"`
		GroupDescription            string    `json:",omitempty"`
		InstanceCapacity_10xlarge   string    `json:",omitempty"`
		InstanceCapacity_2xlarge    string    `json:",omitempty"`
		InstanceCapacity_4xlarge    string    `json:",omitempty"`
		InstanceCapacity_8xlarge    string    `json:",omitempty"`
		InstanceCapacity_large      string    `json:",omitempty"`
		InstanceCapacity_medium     string    `json:",omitempty"`
		InstanceCapacity_xlarge     string    `json:",omitempty"`
		InstanceFamily              string    `json:",omitempty"`
		InstanceType                string    `json:",omitempty"`
		IO                          string    `json:",omitempty"`
		LeaseContractLength         string    `json:",omitempty"`
		LicenseModel                string    `json:",omitempty"`
		Location                    string    `json:",omitempty"`
		LocationType                string    `json:",omitempty"`
		MaxIOPSBurstPerformance     string    `json:",omitempty"`
		MaxIOPSVolume               string    `json:",omitempty"`
		MaxthroughputVolume         string    `json:",omitempty"`
		MaxVolumeSize               int64     `json:",omitempty"`
		Memory                      int64     `json:",omitempty"`
		MinVolumeSize               int64     `json:",omitempty"`
		NetworkPerformance          string    `json:",omitempty"`
		OfferTermCode               string    `json:",omitempty"`
		OperatingSystem             string    `json:",omitempty"`
		Operation                   string    `json:",omitempty"`
		Origin                      string    `json:",omitempty"`
		PhysicalCores               string    `json:",omitempty"`
		PhysicalProcessor           string    `json:",omitempty"`
		PreInstalledSW              string    `json:",omitempty"`
		PriceDescription            string    `json:",omitempty"`
		PricePerUnit                float64   `json:",omitempty"`
		ProcessorArchitecture       string    `json:",omitempty"`
		ProcessorFeatures           string    `json:",omitempty"`
		ProductFamily               string    `json:",omitempty"`
		Provisioned                 bool      `json:",omitempty"`
		PurchaseOption              string    `json:",omitempty"`
		RateCode                    string    `json:",omitempty"`
		Recipient                   string    `json:",omitempty"`
		RelatedTo                   string    `json:",omitempty"`
		RequestDescription          string    `json:",omitempty"`
		RequestType                 string    `json:",omitempty"`
		ResourceEndPoint            string    `json:",omitempty"`
		RoutingTarget               string    `json:",omitempty"`
		RoutingType                 string    `json:",omitempty"`
		ServiceCode                 string    `json:",omitempty"`
		SKU                         string    `json:",omitempty"`
		Sockets                     int       `json:",omitempty"`
		StartingRange               float64   `json:",omitempty"`
		Storage                     string    `json:",omitempty"`
		StorageClass                string    `json:",omitempty"`
		StorageMedia                string    `json:",omitempty"`
		Tenancy                     string    `json:",omitempty"`
		TermType                    string    `json:",omitempty"`
		ToLocation                  string    `json:",omitempty"`
		ToLocationType              string    `json:",omitempty"`
		TransferType                string    `json:",omitempty"`
		Unit                        string    `json:",omitempty"`
		UsageFamily                 string    `json:",omitempty"`
		UsageType                   string    `json:",omitempty"`
		VCPU                        int       `json:",omitempty"`
		VolumeType                  string    `json:",omitempty"`
	}
)

const (
	FormatVersion        = `v1.0`
	BaseURL              = `https://pricing.us-east-1.amazonaws.com`
	AmazonS3URL          = BaseURL + "/offers/v1.0/aws/AmazonS3/current/index.csv"
	AmazonGlacierURL     = BaseURL + "/offers/v1.0/aws/AmazonGlacier/current/index.csv"
	AmazonSESURL         = BaseURL + "/offers/v1.0/aws/AmazonSES/current/index.csv"
	AmazonRDSURL         = BaseURL + "/offers/v1.0/aws/AmazonRDS/current/index.csv"
	AmazonSimpleDBURL    = BaseURL + "/offers/v1.0/aws/AmazonSimpleDB/current/index.csv"
	AmazonDynamoDBURL    = BaseURL + "/offers/v1.0/aws/AmazonDynamoDB/current/index.csv"
	AmazonEC2URL         = BaseURL + "/offers/v1.0/aws/AmazonEC2/current/index.csv"
	AmazonRoute53URL     = BaseURL + "/offers/v1.0/aws/AmazonRoute53/current/index.csv"
	AmazonRedshiftURL    = BaseURL + "/offers/v1.0/aws/AmazonRedshift/current/index.csv"
	AmazonElastiCacheURL = BaseURL + "/offers/v1.0/aws/AmazonElastiCache/current/index.csv"
	AmazonCloudFrontURL  = BaseURL + "/offers/v1.0/aws/AmazonCloudFront/current/index.csv"
	AWSKMSURL            = BaseURL + "/offers/v1.0/aws/awskms/current/index.csv"
	AmazonVPCURL         = BaseURL + "/offers/v1.0/aws/AmazonVPC/current/index.csv"

	AmazonS3OfferCode          = `AmazonS3`
	AmazonGlacierOfferCode     = `AmazonGlacier`
	AmazonSESOfferCode         = `AmazonSES`
	AmazonRDSOfferCode         = `AmazonRDS`
	AmazonSimpleDBOfferCode    = `AmazonSimpleDB`
	AmazonDynamoDBOfferCode    = `AmazonDynamoDB`
	AmazonEC2OfferCode         = `AmazonEC2`
	AmazonRoute53OfferCode     = `AmazonRoute53`
	AmazonRedshiftOfferCode    = `AmazonRedshift`
	AmazonElastiCacheOfferCode = `AmazonElastiCache`
	AmazonCloudFrontOfferCode  = `AmazonCloudFront`
	AWSKMSOfferCode            = `awskms`
	AmazonVPCOfferCode         = `AmazonVPC`

	AsiaPacificRegionFullName          = `Asia Pacific`
	AsiaPacificSeoulRegionFullName     = `Asia Pacific (Seoul)`
	AsiaPacificSingaporeRegionFullName = `Asia Pacific (Singapore)`
	AsiaPacificSydneyRegionFullName    = `Asia Pacific (Sydney)`
	AsiaPacificTokyoRegionFullName     = `Asia Pacific (Tokyo)`
	AustraliaRegionFullName            = `Australia`
	EUFrankfurtRegionFullName          = `EU (Frankfurt)`
	EUIrelandRegionFullName            = `EU (Ireland)`
	EuropeRegionFullName               = `Europe`
	IndiaRegionFullName                = `India`
	JapanRegionFullName                = `Japan`
	SouthAmericaRegionFullName         = `South America`
	SouthAmericaSaoPauloRegionFullName = `South America (Sao Paulo)`
	UnitedStatesRegionFullName         = `United States`
	USEast1RegionFullName              = `US East (N. Virginia)`
	USGovCloudRegionFullName           = `AWS GovCloud (US)`
	USWest1RegionFullName              = `US West (N. California)`
	USWest2RegionFullName              = `US West (Oregon)`

	AsiaPacificRegion          = ``
	AsiaPacificSeoulRegion     = `ap-northeast-2`
	AsiaPacificSingaporeRegion = `ap-southeast-1`
	AsiaPacificSydneyRegion    = `ap-southeast-2`
	AsiaPacificTokyoRegion     = `ap-northeast-1`
	AustraliaRegion            = ``
	EUFrankfurtRegion          = `eu-central-1`
	EUIrelandRegion            = `eu-west-1`
	EuropeRegion               = ``
	IndiaRegion                = ``
	JapanRegion                = ``
	SouthAmericaRegion         = ``
	SouthAmericaSaoPauloRegion = `sa-east-1`
	UnitedStatesRegion         = ``
	USEast1Region              = `us-east-1`
	USGovCloudRegion           = `us-gov-west-1`
	USWest1Region              = `us-west-1`
	USWest2Region              = `us-west-2`

	OnDemandTermType = `OnDemand`
	ReservedTermType = `Reserved`

	DedicatedTenancy = `Dedicated`
	SharedTenancy    = `Shared`
	HostTenancy      = `Host`

	WindowsPlatform = `Windows`
	RHELPlatform    = `RHEL`
	SUSEPlatform    = `SUSE`
	LinuxPlatform   = `Linux`

	AmazonAuroraDatabaseEngine = `Amazon Aurora`
	MariaDBDatabaseEngine      = `MariaDB`
	MySQLDatabaseEngine        = `MySQL`
	OracleDatabaseEngine       = `Oracle`
	PostgreSQLDatabaseEngine   = `PostgreSQL`
	SQLServerDatabaseEngine    = `SQL Server`

	RedisCacheEngine     = `Redis`
	MemcachedCacheEngine = `Memcached`

	SingleAZDeploymentOption               = `Single-AZ`
	MultiAZDeploymentOption                = `Multi-AZ`
	MultiAZSQLServerMirrorDeploymentOption = `Multi-AZ (SQL Server Mirror)`
)

var (
	offersEndpoints = map[string]string{
		AmazonS3OfferCode:          AmazonS3URL,
		AmazonGlacierOfferCode:     AmazonGlacierURL,
		AmazonSESOfferCode:         AmazonSESURL,
		AmazonRDSOfferCode:         AmazonRDSURL,
		AmazonSimpleDBOfferCode:    AmazonSimpleDBURL,
		AmazonDynamoDBOfferCode:    AmazonDynamoDBURL,
		AmazonEC2OfferCode:         AmazonEC2URL,
		AmazonRoute53OfferCode:     AmazonRoute53URL,
		AmazonRedshiftOfferCode:    AmazonRedshiftURL,
		AmazonElastiCacheOfferCode: AmazonElastiCacheURL,
		AmazonCloudFrontOfferCode:  AmazonCloudFrontURL,
		AWSKMSOfferCode:            AWSKMSURL,
		AmazonVPCOfferCode:         AmazonVPCURL,
	}
	regionTranslation = map[string]string{
		AsiaPacificRegionFullName:          AsiaPacificRegion,
		AsiaPacificSeoulRegionFullName:     AsiaPacificSeoulRegion,
		AsiaPacificSingaporeRegionFullName: AsiaPacificSingaporeRegion,
		AsiaPacificSydneyRegionFullName:    AsiaPacificSydneyRegion,
		AsiaPacificTokyoRegionFullName:     AsiaPacificTokyoRegion,
		AustraliaRegionFullName:            AustraliaRegion,
		EUFrankfurtRegionFullName:          EUFrankfurtRegion,
		EUIrelandRegionFullName:            EUIrelandRegion,
		EuropeRegionFullName:               EuropeRegion,
		IndiaRegionFullName:                IndiaRegion,
		JapanRegionFullName:                JapanRegion,
		SouthAmericaRegionFullName:         SouthAmericaRegion,
		SouthAmericaSaoPauloRegionFullName: SouthAmericaSaoPauloRegion,
		UnitedStatesRegionFullName:         UnitedStatesRegion,
		USEast1RegionFullName:              USEast1Region,
		USGovCloudRegionFullName:           USGovCloudRegion,
		USWest1RegionFullName:              USWest1Region,
		USWest2RegionFullName:              USWest2Region,
	}
)

var dataDir = `.pricing`

func LoadTable() (*Table, error) {

	start := time.Now()
	fmt.Fprint(os.Stderr, start.Format("2006-01-02 15:04:05"), ": loading pricing tables... ")
	defer func() { fmt.Fprint(os.Stderr, time.Since(start), "\n") }()

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("unable to make dataDir: %q", dataDir)
	}

	table := &Table{
		Rows: make([]*Row, 0, 128<<10),
	}
	table_me := sync.Mutex{}

	errors := make(chan error, len(offersEndpoints))

	for offerCode, endpoint := range offersEndpoints {
		go func(offerCode, endpoint string) {
			rc, err := get(endpoint)
			if err != nil {
				errors <- err
				return
			}
			defer rc.Close()

			r := csv.NewReader(rc)

			// discard
			r.FieldsPerRecord = 2
			r.Read() // "FormatVersion","v1.0"
			r.Read() // "Disclaimer","This pricing list is for informational purposes only. All prices are subject to the additional terms included in the pricing pages on http://aws.amazon.com. All Free Tier prices are also subject to the terms included at https://aws.amazon.com/free/"
			r.Read() // "Publication Date","2016-01-07T02:15:22Z"
			r.Read() // "Version","20151201000000"
			r.Read() // "OfferCode","AmazonVPC"

			r.FieldsPerRecord = 0
			fields, err := r.Read()
			if err != nil {
				errors <- err
				return
			}

			var rows []*Row
			for {
				line, err := r.Read()
				if err != nil {
					if err == io.EOF {
						table_me.Lock()
						table.Rows = append(table.Rows, rows...)
						table_me.Unlock()
						errors <- nil
					} else {
						errors <- err
					}
					return
				}
				row := &Row{}
				for i, field := range fields {
					if len(line[i]) == 0 {
						continue
					}
					switch field {
					case "Availability":
						row.Availability = line[i]
					case "Cache Engine":
						row.CacheEngine = line[i]
					case "Clock Speed":
						// ClockSpeed: 2.9 GHz,2.6 GHz,2.8 GHz,2.4  GHz,2.5 GHz,Up to 3.3 GHz,Up to 3.0 GHz,2 GHz,,2.4 GHz,
						row.ClockSpeed, _ = strconv.ParseFloat(
							strings.TrimSuffix(strings.TrimPrefix(line[i], "Up to "), " GHz"),
							64,
						)
					case "Content Type":
						row.ContentType = line[i]
					case "Currency":
						row.Currency = line[i]
					case "Current Generation":
						row.CurrentGeneration = line[i] == "Yes"
					case "Database Edition":
						row.DatabaseEdition = line[i]
					case "Database Engine":
						row.DatabaseEngine = line[i]
					case "Dedicated EBS Throughput":
						row.DedicatedEBSThroughput = line[i]
					case "Deployment Option":
						row.DeploymentOption = line[i]
					case "Description":
						row.Description = line[i]
					case "Durability":
						row.Durability = line[i]
					case "EBS Optimized":
						row.EBSOptimized = line[i] == "Yes"
					case "ECU":
						row.ECU, _ = strconv.Atoi(line[i])
					case "EffectiveDate":
						row.EffectiveDate, _ = time.Parse("2006-01-02", line[i])
					case "EndingRange":
						if v := line[i]; v == "Inf" {
							row.EndingRange = math.Inf(1)
						} else {
							row.EndingRange, _ = strconv.ParseFloat(v, 64)
						}
					case "Endpoint Type":
						row.EndpointType = line[i]
					case "Engine Code":
						row.EngineCode = line[i]
					case "Enhanced Networking Supported":
						row.EnhancedNetworkingSupported = line[i] == "Yes"
					case "Fee Code":
						row.FeeCode = line[i]
					case "Fee Description":
						row.FeeDescription = line[i]
					case "From Location Type":
						row.FromLocationType = line[i]
					case "From Location":
						row.FromLocation = line[i]
					case "GPU":
						row.GPU, _ = strconv.Atoi(line[i])
					case "Group Description":
						row.GroupDescription = line[i]
					case "Group":
						row.Group = line[i]
					case "I/O":
						row.IO = line[i]
					case "Instance Capacity - 10xlarge":
						row.InstanceCapacity_10xlarge = line[i]
					case "Instance Capacity - 2xlarge":
						row.InstanceCapacity_2xlarge = line[i]
					case "Instance Capacity - 4xlarge":
						row.InstanceCapacity_4xlarge = line[i]
					case "Instance Capacity - 8xlarge":
						row.InstanceCapacity_8xlarge = line[i]
					case "Instance Capacity - large":
						row.InstanceCapacity_large = line[i]
					case "Instance Capacity - medium":
						row.InstanceCapacity_medium = line[i]
					case "Instance Capacity - xlarge":
						row.InstanceCapacity_xlarge = line[i]
					case "Instance Family":
						row.InstanceFamily = line[i]
					case "Instance Type":
						row.InstanceType = line[i]
					case "LeaseContractLength":
						row.LeaseContractLength = line[i]
					case "License Model":
						row.LicenseModel = line[i]
					case "Location Type":
						row.LocationType = line[i]
					case "Location":
						row.Location = line[i]
						row.Region = regionTranslation[row.Location]
					case "Max IOPS Burst Performance":
						row.MaxIOPSBurstPerformance = line[i]
					case "Max IOPS/volume":
						row.MaxIOPSVolume = line[i]
					case "Max Volume Size":
						// MaxVolumeSize: 16 TiB,1 TiB,,3 TB,6 TB,64 TB,
						row.MaxVolumeSize = parseHumanReadableSize(line[i])
					case "Max throughput/volume":
						row.MaxthroughputVolume = line[i]
					case "Memory":
						// 4 GiB,34.2 GiB,6.6 GiB,2.78 GiB,3.75 GiB,8 GiB,64 GiB,1 GiB,61 GiB,117 GiB,120 GiB,244 GiB,3.22 GiB,16.7 GiB,0.213 GiB,23 GiB,16 GiB,0.613 GiB,118 GiB,160 GiB,1.7 GiB,,31 GiB,13.3 GiB,28.4 GiB,6.05 GiB,60.5 GiB,30.5 GiB,60 GiB,7 GiB,33.8 GiB,0.555 GiB,237 GiB,2 GiB,7.5 GiB,22.5 GiB,68.4 GiB,14.6 GiB,27.9 GiB,7.1 GiB,1.55 GiB,15.25 GiB,15 GiB,58.2 GiB,13.5 GiB,32 GiB,17.1 GiB,1.3 GiB,68 GiB,3.35 GiB,122 GiB,30 GiB,
						row.Memory = parseHumanReadableSize(line[i])
					case "Min Volume Size":
						// MinVolumeSize: ,5 GB,100 GB,10 GB,
						row.MinVolumeSize = parseHumanReadableSize(line[i])
					case "Network Performance":
						row.NetworkPerformance = line[i]
					case "OfferTermCode":
						row.OfferTermCode = line[i]
					case "Operating System":
						row.OperatingSystem = line[i]
					case "Origin":
						row.Origin = line[i]
					case "Physical Cores":
						row.PhysicalCores = line[i]
					case "Physical Processor":
						row.PhysicalProcessor = line[i]
					case "Pre Installed S/W":
						row.PreInstalledSW = line[i]
					case "PriceDescription":
						row.PriceDescription = line[i]
					case "PricePerUnit":
						row.PricePerUnit, _ = strconv.ParseFloat(line[i], 64)
					case "Processor Architecture":
						row.ProcessorArchitecture = line[i]
					case "Processor Features":
						row.ProcessorFeatures = line[i]
					case "Product Family":
						row.ProductFamily = line[i]
					case "Provisioned":
						row.Provisioned = line[i] == "Yes"
					case "PurchaseOption":
						row.PurchaseOption = line[i]
					case "RateCode":
						row.RateCode = line[i]
					case "Recipient":
						row.Recipient = line[i]
					case "RelatedTo":
						row.RelatedTo = line[i]
					case "Request Description":
						row.RequestDescription = line[i]
					case "Request Type":
						row.RequestType = line[i]
					case "Resource EndPoint":
						row.ResourceEndPoint = line[i]
					case "Routing Target":
						row.RoutingTarget = line[i]
					case "Routing Type":
						row.RoutingType = line[i]
					case "SKU":
						row.SKU = line[i]
					case "Sockets":
						row.Sockets, _ = strconv.Atoi(line[i])
					case "StartingRange":
						if v := line[i]; v == "Inf" {
							row.StartingRange = math.Inf(1)
						} else {
							row.StartingRange, _ = strconv.ParseFloat(v, 64)
						}
					case "Storage Class":
						row.StorageClass = line[i]
					case "Storage Media":
						row.StorageMedia = line[i]
					case "Storage":
						row.Storage = line[i]
					case "Tenancy":
						row.Tenancy = line[i]
					case "TermType":
						row.TermType = line[i]
					case "To Location Type":
						row.ToLocationType = line[i]
					case "To Location":
						row.ToLocation = line[i]
					case "Transfer Type":
						row.TransferType = line[i]
					case "Unit":
						row.Unit = line[i]
					case "Usage Family":
						row.UsageFamily = line[i]
					case "Volume Type":
						row.VolumeType = line[i]
					case "operation":
						row.Operation = line[i]
					case "serviceCode":
						row.OfferCode = line[i]
					case "usageType":
						row.UsageType = line[i]
					case "vCPU":
						row.VCPU, _ = strconv.Atoi(line[i])
					}
				}
				rows = append(rows, row)
			}
		}(offerCode, endpoint)
	}

	for i := 0; i < cap(errors); i++ {
		if err := <-errors; err != nil {
			return nil, err
		}
	}

	return table, nil

}
