package window

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/emptyinterface/window/pricing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

type (
	DBInstance struct {
		// Specifies the allocated storage size specified in gigabytes.
		AllocatedStorage int64

		// Indicates that minor version patches are applied automatically.
		AutoMinorVersionUpgrade bool

		// Specifies the name of the Availability Zone the DB instance is located in.
		AvailabilityZoneName string

		// Specifies the number of days for which automatic DB snapshots are retained.
		BackupRetentionPeriod int64

		// The identifier of the CA certificate for this DB instance.
		CACertificateIdentifier string

		// If present, specifies the name of the character set that this instance is
		// associated with.
		CharacterSetName string

		// Specifies whether tags are copied from the DB instance to snapshots of the
		// DB instance.
		CopyTagsToSnapshot bool

		// If the DB instance is a member of a DB cluster, contains the name of the
		// DB cluster that the DB instance is a member of.
		DBClusterIdentifier string

		// Contains the name of the compute and memory capacity class of the DB instance.
		DBInstanceClass string

		// Contains a user-supplied database identifier. This identifier is the unique
		// key that identifies a DB instance.
		DBInstanceIdentifier string

		// Specifies the current state of this database.
		DBInstanceStatus string

		// The meaning of this parameter differs according to the database engine you
		// use. For example, this value returns either MySQL or PostgreSQL information
		// when returning values from CreateDBInstanceReadReplica since Read Replicas
		// are only supported for MySQL and PostgreSQL.
		//
		//  MySQL, SQL Server, PostgreSQL, Amazon Aurora
		//
		//  Contains the name of the initial database of this instance that was provided
		// at create time, if one was specified when the DB instance was created. This
		// same name is returned for the life of the DB instance.
		//
		// Type: String
		//
		//  Oracle
		//
		//  Contains the Oracle System ID (SID) of the created DB instance. Not shown
		// when the returned parameters do not apply to an Oracle DB instance.
		DBName string

		// Provides the list of DB parameter groups applied to this DB instance.
		DBParameterGroups []*rds.DBParameterGroupStatus

		// Provides List of DB security group elements containing only DBSecurityGroup.Name
		// and DBSecurityGroup.Status subelements.
		DBSecurityGroups []*rds.DBSecurityGroupMembership

		// Specifies information on the subnet group associated with the DB instance,
		// including the name, description, and subnets in the subnet group.
		DBSubnetGroup *rds.DBSubnetGroup

		// Specifies the port that the DB instance listens on. If the DB instance is
		// part of a DB cluster, this can be a different port than the DB cluster port.
		DbInstancePort int64

		// If StorageEncrypted is true, the region-unique, immutable identifier for
		// the encrypted DB instance. This identifier is found in AWS CloudTrail Log
		// entries whenever the KMS key for the DB instance is accessed.
		DbiResourceId string

		// Specifies the connection endpoint.
		Endpoint *rds.Endpoint

		// Provides the name of the database engine to be used for this DB instance.
		Engine string

		// Indicates the database engine version.
		EngineVersion string

		// Provides the date and time the DB instance was created.
		InstanceCreateTime time.Time

		// Specifies the Provisioned IOPS (I/O operations per second) value.
		Iops int64

		// If StorageEncrypted is true, the KMS key identifier for the encrypted DB
		// instance.
		KmsKeyId string

		// Specifies the latest time to which a database can be restored with point-in-time
		// restore.
		LatestRestorableTime time.Time

		// License model information for this DB instance.
		LicenseModel string

		// Contains the master username for the DB instance.
		MasterUsername string

		// Specifies if the DB instance is a Multi-AZ deployment.
		MultiAZ bool

		// Provides the list of option group memberships for this DB instance.
		OptionGroupMemberships []*rds.OptionGroupMembership

		// Specifies that changes to the DB instance are pending. This element is only
		// included when changes are pending. Specific changes are identified by subelements.
		PendingModifiedValues *rds.PendingModifiedValues

		// Specifies the daily time range during which automated backups are created
		// if automated backups are enabled, as determined by the BackupRetentionPeriod.
		PreferredBackupWindow string

		// Specifies the weekly time range during which system maintenance can occur,
		// in Universal Coordinated Time (UTC).
		PreferredMaintenanceWindow string

		// Specifies the accessibility options for the DB instance. A value of true
		// specifies an Internet-facing instance with a publicly resolvable DNS name,
		// which resolves to a public IP address. A value of false specifies an internal
		// instance with a DNS name that resolves to a private IP address.
		//
		//  Default: The default behavior varies depending on whether a VPC has been
		// requested or not. The following list shows the default behavior in each case.
		//
		//   Default VPC:true  VPC:false   If no DB subnet group has been specified
		// as part of the request and the PubliclyAccessible value has not been set,
		// the DB instance will be publicly accessible. If a specific DB subnet group
		// has been specified as part of the request and the PubliclyAccessible value
		// has not been set, the DB instance will be private.
		PubliclyAccessible bool

		// Contains one or more identifiers of the Read Replicas associated with this
		// DB instance.
		ReadReplicaDBInstanceIdentifiers []string

		// Contains the identifier of the source DB instance if this DB instance is
		// a Read Replica.
		ReadReplicaSourceDBInstanceIdentifier string

		// If present, specifies the name of the secondary Availability Zone for a DB
		// instance with multi-AZ support.
		SecondaryAvailabilityZone string

		// The status of a Read Replica. If the instance is not a Read Replica, this
		// will be blank.
		StatusInfos []*rds.DBInstanceStatusInfo

		// Specifies whether the DB instance is encrypted.
		StorageEncrypted bool

		// Specifies the storage type associated with DB instance.
		StorageType string

		// The ARN from the Key Store with which the instance is associated for TDE
		// encryption.
		TdeCredentialArn string

		// Provides List of VPC security group elements that the DB instance belongs
		// to.
		VpcSecurityGroups []*rds.VpcSecurityGroupMembership

		Name             string
		Id               string
		State            string
		Region           *Region
		Classic          *Classic
		VPC              *VPC
		AvailabilityZone *AvailabilityZone
		StorageTypeName  string
		SecurityGroups   []*SecurityGroup
		CloudWatchAlarms []*CloudWatchAlarm

		MemoryCapacity int64
		Stats          *DBInstanceStats

		Log []string
	}

	DBInstanceByNameAsc                  []*DBInstance
	DescribeLogsDetailsByLastWrittenDesc []*rds.DescribeDBLogFilesDetails

	DBInstanceSet []*DBInstance
)

const RDSMaxLogLines = 50

func (a DBInstanceByNameAsc) Len() int      { return len(a) }
func (a DBInstanceByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DBInstanceByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (a DescribeLogsDetailsByLastWrittenDesc) Len() int      { return len(a) }
func (a DescribeLogsDetailsByLastWrittenDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DescribeLogsDetailsByLastWrittenDesc) Less(i, j int) bool {
	return aws.Int64Value(a[i].LastWritten) > aws.Int64Value(a[j].LastWritten)
}

func LoadDBInstances(input *rds.DescribeDBInstancesInput) (map[string]*DBInstance, error) {

	dbs := map[string]*DBInstance{}

	if err := RDSClient.DescribeDBInstancesPages(input, func(page *rds.DescribeDBInstancesOutput, _ bool) bool {
		for _, rdsDbinst := range page.DBInstances {
			dbinst := &DBInstance{
				AllocatedStorage:                      aws.Int64Value(rdsDbinst.AllocatedStorage),
				AutoMinorVersionUpgrade:               aws.BoolValue(rdsDbinst.AutoMinorVersionUpgrade),
				AvailabilityZoneName:                  aws.StringValue(rdsDbinst.AvailabilityZone),
				BackupRetentionPeriod:                 aws.Int64Value(rdsDbinst.BackupRetentionPeriod),
				CACertificateIdentifier:               aws.StringValue(rdsDbinst.CACertificateIdentifier),
				CharacterSetName:                      aws.StringValue(rdsDbinst.CharacterSetName),
				CopyTagsToSnapshot:                    aws.BoolValue(rdsDbinst.CopyTagsToSnapshot),
				DBClusterIdentifier:                   aws.StringValue(rdsDbinst.DBClusterIdentifier),
				DBInstanceClass:                       aws.StringValue(rdsDbinst.DBInstanceClass),
				DBInstanceIdentifier:                  aws.StringValue(rdsDbinst.DBInstanceIdentifier),
				DBInstanceStatus:                      aws.StringValue(rdsDbinst.DBInstanceStatus),
				DBName:                                aws.StringValue(rdsDbinst.DBName),
				DBParameterGroups:                     rdsDbinst.DBParameterGroups,
				DBSecurityGroups:                      rdsDbinst.DBSecurityGroups,
				DBSubnetGroup:                         rdsDbinst.DBSubnetGroup,
				DbInstancePort:                        aws.Int64Value(rdsDbinst.DbInstancePort),
				DbiResourceId:                         aws.StringValue(rdsDbinst.DbiResourceId),
				Endpoint:                              rdsDbinst.Endpoint,
				Engine:                                aws.StringValue(rdsDbinst.Engine),
				EngineVersion:                         aws.StringValue(rdsDbinst.EngineVersion),
				InstanceCreateTime:                    aws.TimeValue(rdsDbinst.InstanceCreateTime),
				Iops:                                  aws.Int64Value(rdsDbinst.Iops),
				KmsKeyId:                              aws.StringValue(rdsDbinst.KmsKeyId),
				LatestRestorableTime:                  aws.TimeValue(rdsDbinst.LatestRestorableTime),
				LicenseModel:                          aws.StringValue(rdsDbinst.LicenseModel),
				MasterUsername:                        aws.StringValue(rdsDbinst.MasterUsername),
				MultiAZ:                               aws.BoolValue(rdsDbinst.MultiAZ),
				OptionGroupMemberships:                rdsDbinst.OptionGroupMemberships,
				PendingModifiedValues:                 rdsDbinst.PendingModifiedValues,
				PreferredBackupWindow:                 aws.StringValue(rdsDbinst.PreferredBackupWindow),
				PreferredMaintenanceWindow:            aws.StringValue(rdsDbinst.PreferredMaintenanceWindow),
				PubliclyAccessible:                    aws.BoolValue(rdsDbinst.PubliclyAccessible),
				ReadReplicaDBInstanceIdentifiers:      aws.StringValueSlice(rdsDbinst.ReadReplicaDBInstanceIdentifiers),
				ReadReplicaSourceDBInstanceIdentifier: aws.StringValue(rdsDbinst.ReadReplicaSourceDBInstanceIdentifier),
				SecondaryAvailabilityZone:             aws.StringValue(rdsDbinst.SecondaryAvailabilityZone),
				StatusInfos:                           rdsDbinst.StatusInfos,
				StorageEncrypted:                      aws.BoolValue(rdsDbinst.StorageEncrypted),
				StorageType:                           aws.StringValue(rdsDbinst.StorageType),
				TdeCredentialArn:                      aws.StringValue(rdsDbinst.TdeCredentialArn),
				VpcSecurityGroups:                     rdsDbinst.VpcSecurityGroups,
			}
			dbinst.Name = dbinst.DBInstanceIdentifier
			dbinst.Id = "rds:" + dbinst.DBInstanceIdentifier
			dbinst.State = dbinst.DBInstanceStatus
			switch dbinst.StorageType {
			case "io1":
				dbinst.StorageTypeName = "IOPS"
			case "gp2":
				dbinst.StorageTypeName = "SSD"
			case "standard":
				dbinst.StorageTypeName = "DISK"
			default:
				dbinst.StorageTypeName = dbinst.StorageType
			}
			dbs[dbinst.DBInstanceIdentifier] = dbinst
		}
		return true
	}); err != nil {
		return nil, err
	}

	return dbs, nil

}

func (dbinst *DBInstance) priceKey() string {
	var option string
	if dbinst.MultiAZ {
		option = pricing.MultiAZDeploymentOption
	} else {
		option = pricing.SingleAZDeploymentOption
	}
	var engine string
	switch dbinst.Engine {
	case "amazonaurora":
		engine = pricing.AmazonAuroraDatabaseEngine
	case "mariadb":
		engine = pricing.MariaDBDatabaseEngine
	case "mysql":
		engine = pricing.MySQLDatabaseEngine
	case "oracle":
		engine = pricing.OracleDatabaseEngine
	case "postgres":
		engine = pricing.PostgreSQLDatabaseEngine
	case "sqlserver":
		engine = pricing.SQLServerDatabaseEngine
	}
	key := fmt.Sprintf("%s:%s:%s:%s:%s",
		pricing.AmazonRDSOfferCode,
		pricing.OnDemandTermType,
		option,
		dbinst.DBInstanceClass,
		engine,
	)
	return key
}

func (dbinst *DBInstance) HourlyCost() float64 {
	if offer, exists := dbinst.Region.Prices[dbinst.priceKey()]; exists {
		return offer.PricePerUnit
	}
	fmt.Println("miss", dbinst.priceKey())
	return 0
}

func (dbinst *DBInstance) MonthlyCost() float64 {
	if offer, exists := dbinst.Region.Prices[dbinst.priceKey()]; exists {
		return offer.PricePerUnit * 24 * 30
	}
	fmt.Println("miss", dbinst.priceKey())
	return 0
}

func (db *DBInstance) Poll() []chan error {

	var errs []chan error

	db.Stats = &DBInstanceStats{}

	errs = append(errs, db.Region.Throttle.do(db.Name+" LOG POLL", func() error {
		resp, err := RDSClient.DescribeDBLogFiles(&rds.DescribeDBLogFilesInput{
			DBInstanceIdentifier: aws.String(db.DBInstanceIdentifier),
			FileSize:             aws.Int64(1),
		})
		if err != nil {
			return err
		}

		sort.Sort(DescribeLogsDetailsByLastWrittenDesc(resp.DescribeDBLogFiles))

		db.Log = nil
		for len(db.Log) < RDSMaxLogLines && len(resp.DescribeDBLogFiles) > 0 {
			details := resp.DescribeDBLogFiles[0]
			resp.DescribeDBLogFiles = resp.DescribeDBLogFiles[1:]
			resp, err := RDSClient.DownloadDBLogFilePortion(&rds.DownloadDBLogFilePortionInput{
				DBInstanceIdentifier: aws.String(db.DBInstanceIdentifier),
				LogFileName:          details.LogFileName,
				NumberOfLines:        aws.Int64(RDSMaxLogLines - int64(len(db.Log))),
			})
			if err != nil {
				return err
			}
			if resp.LogFileData != nil {
				if line := strings.TrimSpace(*resp.LogFileData); len(line) > 0 {
					// preserve order from oldest (top) to newest (bottom) entries
					db.Log = append(strings.Split(*resp.LogFileData, "\n"), db.Log...)
				}
			}
		}
		return nil
	}))

	for _, m := range RDSMetrics {
		m := m
		errs = append(errs, db.Region.Throttle.do(db.Name+" METRICS POLL", func() error {
			return m.RunFor(db)
		}))
	}

	return errs

}

func (db *DBInstance) Inactive() bool {
	return false
}

func (db *DBInstance) DiskInUse() float64 {
	if db.Stats != nil {
		total := db.AllocatedStorage << 30 // gigabytes conversion
		return (float64(total-db.Stats.FreeStorageSpace) / float64(total)) * 100
	}
	return 0
}

func (dbs DBInstanceSet) Summary() *DBInstanceStats {

	stats := &DBInstanceStats{}

	var has_stats bool
	for _, db := range dbs {
		if db.Stats != nil {
			has_stats = true
			stats.CPUUtilization += db.Stats.CPUUtilization
			stats.DatabaseConnections += db.Stats.DatabaseConnections
			stats.ReadIOPS += db.Stats.ReadIOPS
			stats.WriteIOPS += db.Stats.WriteIOPS
			if db.Stats.ReadLatency.Min < stats.ReadLatency.Min || stats.ReadLatency.Min == 0 {
				stats.ReadLatency.Min = db.Stats.ReadLatency.Min
			}
			if db.Stats.ReadLatency.Max > stats.ReadLatency.Max || stats.ReadLatency.Max == 0 {
				stats.ReadLatency.Max = db.Stats.ReadLatency.Max
			}
			stats.ReadLatency.Avg += db.Stats.ReadLatency.Avg
			if db.Stats.WriteLatency.Min < stats.WriteLatency.Min || stats.WriteLatency.Min == 0 {
				stats.WriteLatency.Min = db.Stats.WriteLatency.Min
			}
			if db.Stats.WriteLatency.Max > stats.WriteLatency.Max || stats.WriteLatency.Max == 0 {
				stats.WriteLatency.Max = db.Stats.WriteLatency.Max
			}
			stats.WriteLatency.Avg += db.Stats.WriteLatency.Avg
			stats.DiskReadThroughputBytesPerSecond += db.Stats.DiskReadThroughputBytesPerSecond
			stats.DiskWriteThroughputBytesPerSecond += db.Stats.DiskWriteThroughputBytesPerSecond
			stats.NetworkReceiveThroughputBytesPerSecond += db.Stats.NetworkReceiveThroughputBytesPerSecond
			stats.NetworkTransmitThroughputBytesPerSecond += db.Stats.NetworkTransmitThroughputBytesPerSecond
		}
	}

	if !has_stats {
		return nil
	}

	stats.CPUUtilization /= float64(len(dbs))
	stats.ReadLatency.Min = roundTime(stats.ReadLatency.Min)
	stats.ReadLatency.Max = roundTime(stats.ReadLatency.Max)
	stats.ReadLatency.Avg = roundTime(stats.ReadLatency.Avg / time.Duration(len(dbs)))
	stats.WriteLatency.Min = roundTime(stats.WriteLatency.Min)
	stats.WriteLatency.Max = roundTime(stats.WriteLatency.Max)
	stats.WriteLatency.Avg = roundTime(stats.WriteLatency.Avg / time.Duration(len(dbs)))

	return stats

}
