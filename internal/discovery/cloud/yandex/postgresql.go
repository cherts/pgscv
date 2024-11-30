package yandex

import (
	"context"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

// Database struct
type Database struct {
	Name  string
	Owner string
}

// Host struct
type Host struct {
	Name   string
	ZoneID string
	Role   postgresql.Host_Role
	Health postgresql.Host_Health
}

// Cluster struct
type Cluster struct {
	ID               string
	Name             string
	FolderID         string
	Health           postgresql.Cluster_Health
	Status           postgresql.Cluster_Status
	ResourcePresetID string
	DiskTypeID       string
	DiskSize         int64
	Hosts            []Host
	Databases        []Database
}

// GetPostgreSQLClusters get a filtered list of clusters and their databases from Yandex cloud API
func (sdk *SDK) GetPostgreSQLClusters(ctx context.Context, folderID string, filter []Filter) ([]Cluster, error) {
	yandexSdk, err := sdk.Build(ctx)
	if err != nil {
		return nil, err
	}

	var req postgresql.ListClustersRequest
	req.FolderId = folderID
	resp, err := yandexSdk.MDB().PostgreSQL().Cluster().List(ctx, &req)
	if err != nil {
		return nil, err
	}
	var clusters []Cluster
	for _, cluster := range resp.Clusters {
		if !(cluster.Status == postgresql.Cluster_RUNNING || cluster.Status == postgresql.Cluster_UPDATING) {
			continue
		}
		matched := make([]int, 0)
		for c, filterCluster := range filter {
			if !filterCluster.MatchName(cluster.Name) {
				continue
			}
			matched = append(matched, c)
		}
		if len(matched) == 0 {
			continue
		}
		var hosts []Host
		var databases []Database
		hostsIterator := yandexSdk.MDB().PostgreSQL().Cluster().ClusterHostsIterator(ctx,
			&postgresql.ListClusterHostsRequest{ClusterId: cluster.Id})
		for hostsIterator.Next() {
			host := hostsIterator.Value()
			if host.Health != postgresql.Host_ALIVE {
				continue
			}
			hosts = append(hosts, Host{
				Name:   host.Name,
				ZoneID: host.ZoneId,
				Role:   host.Role,
				Health: host.Health,
			})
		}
		dbResp, err := yandexSdk.MDB().PostgreSQL().Database().List(ctx,
			&postgresql.ListDatabasesRequest{ClusterId: cluster.Id})
		if err != nil {
			return nil, err
		}
		for _, database := range dbResp.Databases {
			skip := true
			for _, f := range matched {
				if !filter[f].MatchDb(database.Name) {
					continue
				}
				skip = false
				break
			}
			if !skip {
				databases = append(databases, Database{
					Name:  database.Name,
					Owner: database.Owner,
				})
			}
		}
		if len(databases) == 0 {
			continue
		}
		clusters = append(clusters, Cluster{
			ID:               cluster.Id,
			Name:             cluster.Name,
			FolderID:         cluster.FolderId,
			Health:           cluster.Health,
			Status:           cluster.Status,
			ResourcePresetID: cluster.Config.Resources.ResourcePresetId,
			DiskTypeID:       cluster.Config.Resources.DiskTypeId,
			DiskSize:         cluster.Config.Resources.DiskSize,
			Hosts:            hosts,
			Databases:        databases,
		})
	}
	return clusters, nil
}
