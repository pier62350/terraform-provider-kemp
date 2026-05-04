// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// GSLBCluster represents a remote LoadMaster pool cluster.
type GSLBCluster struct {
	IP      string `json:"IP"`
	Name    string `json:"Name"`
	Type    string `json:"Type,omitempty"`
	LatSecs int32  `json:"LatSecs,omitempty"`
	LonSecs int32  `json:"LonSecs,omitempty"`
}

type gslbClusterListResponse struct {
	Response
	Cluster []GSLBCluster `json:"Cluster,omitempty"`
}

type gslbClusterShowResponse struct {
	Response
	Cluster GSLBCluster `json:"Cluster"`
}

// GSLBFQDNMember is one IP entry in a GSLB FQDN's member list.
type GSLBFQDNMember struct {
	IP      string `json:"IP"`
	Cluster string `json:"Cluster,omitempty"`
	Checker string `json:"Checker,omitempty"`
}

// GSLBFQDN is a GSLB fully-qualified domain name with its member list.
type GSLBFQDN struct {
	Name              string           `json:"Name"`
	SelectionCriteria string           `json:"SelectionCriteria,omitempty"`
	FailTime          int32            `json:"FailTime,omitempty"`
	Members           []GSLBFQDNMember `json:"MappedAddress,omitempty"`
}

type gslbFQDNListResponse struct {
	Response
	Fqdn []GSLBFQDN `json:"Fqdn,omitempty"`
}

type gslbFQDNShowResponse struct {
	Response
	Fqdn GSLBFQDN `json:"Fqdn"`
}

type gslbCustomLocationListResponse struct {
	Response
	CustomLocation []string `json:"CustomLocation,omitempty"`
}

// ---- Cluster functions ----

func (c *Client) AddCluster(ctx context.Context, ip, name string) error {
	type body struct {
		IP   string `json:"ip"`
		Name string `json:"name"`
	}
	return c.call(ctx, "addcluster", body{IP: ip, Name: name}, nil)
}

func (c *Client) ModifyCluster(ctx context.Context, ip, name, clusterType string) error {
	type body struct {
		IP          string `json:"ip"`
		Name        string `json:"name,omitempty"`
		ClusterType string `json:"type,omitempty"`
	}
	return c.call(ctx, "modcluster", body{IP: ip, Name: name, ClusterType: clusterType}, nil)
}

func (c *Client) ChangeClusterLocation(ctx context.Context, ip string, latSecs, lonSecs int64) error {
	type body struct {
		IP      string `json:"ip"`
		LatSecs int64  `json:"latsecs"`
		LonSecs int64  `json:"longsecs"`
	}
	return c.call(ctx, "clustchangeloc", body{IP: ip, LatSecs: latSecs, LonSecs: lonSecs}, nil)
}

func (c *Client) ShowCluster(ctx context.Context, ip string) (*GSLBCluster, error) {
	type body struct {
		IP string `json:"ip"`
	}
	var resp gslbClusterShowResponse
	if err := c.call(ctx, "showcluster", body{IP: ip}, &resp); err != nil {
		return nil, err
	}
	return &resp.Cluster, nil
}

func (c *Client) ListClusters(ctx context.Context) ([]GSLBCluster, error) {
	var resp gslbClusterListResponse
	if err := c.call(ctx, "listclusters", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Cluster, nil
}

func (c *Client) DeleteCluster(ctx context.Context, ip string) error {
	type body struct {
		IP string `json:"ip"`
	}
	return c.call(ctx, "delcluster", body{IP: ip}, nil)
}

// ---- Custom location functions ----

func (c *Client) AddCustomLocation(ctx context.Context, name string) error {
	type body struct {
		Location string `json:"location"`
	}
	return c.call(ctx, "addcustomlocation", body{Location: name}, nil)
}

func (c *Client) RenameCustomLocation(ctx context.Context, oldName, newName string) error {
	type body struct {
		OldName string `json:"cloldname"`
		NewName string `json:"clnewname"`
	}
	return c.call(ctx, "editcustomlocation", body{OldName: oldName, NewName: newName}, nil)
}

func (c *Client) DeleteCustomLocation(ctx context.Context, name string) error {
	type body struct {
		Name string `json:"clName"`
	}
	return c.call(ctx, "deletecustomlocation", body{Name: name}, nil)
}

func (c *Client) ListCustomLocations(ctx context.Context) ([]string, error) {
	var resp gslbCustomLocationListResponse
	if err := c.call(ctx, "listcustomlocation", nil, &resp); err != nil {
		return nil, err
	}
	return resp.CustomLocation, nil
}

// ---- FQDN functions ----

func (c *Client) AddFQDN(ctx context.Context, fqdn string) error {
	type body struct {
		FQDN string `json:"fqdn"`
	}
	return c.call(ctx, "addfqdn", body{FQDN: fqdn}, nil)
}

func (c *Client) ModifyFQDN(ctx context.Context, fqdn, selectionCriteria string, failTime int32) error {
	type body struct {
		FQDN              string `json:"fqdn"`
		SelectionCriteria string `json:"SelectionCriteria,omitempty"`
		FailTime          int32  `json:"FailTime,omitempty"`
	}
	return c.call(ctx, "modfqdn", body{FQDN: fqdn, SelectionCriteria: selectionCriteria, FailTime: failTime}, nil)
}

func (c *Client) ShowFQDN(ctx context.Context, fqdn string) (*GSLBFQDN, error) {
	type body struct {
		FQDN string `json:"fqdn"`
	}
	var resp gslbFQDNShowResponse
	if err := c.call(ctx, "showfqdn", body{FQDN: fqdn}, &resp); err != nil {
		return nil, err
	}
	return &resp.Fqdn, nil
}

func (c *Client) ListFQDNs(ctx context.Context) ([]GSLBFQDN, error) {
	var resp gslbFQDNListResponse
	if err := c.call(ctx, "listfqdns", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Fqdn, nil
}

func (c *Client) DeleteFQDN(ctx context.Context, fqdn string) error {
	type body struct {
		FQDN string `json:"fqdn"`
	}
	return c.call(ctx, "delfqdn", body{FQDN: fqdn}, nil)
}

// ---- FQDN member functions ----

func (c *Client) AddFQDNMember(ctx context.Context, fqdn, ip, cluster string) error {
	type body struct {
		FQDN    string `json:"fqdn"`
		IP      string `json:"ip"`
		Cluster string `json:"cluster,omitempty"`
	}
	return c.call(ctx, "addmap", body{FQDN: fqdn, IP: ip, Cluster: cluster}, nil)
}

func (c *Client) ModifyFQDNMember(ctx context.Context, fqdn, ip, checker string) error {
	type body struct {
		FQDN    string `json:"fqdn"`
		IP      string `json:"ip"`
		Checker string `json:"checker,omitempty"`
	}
	return c.call(ctx, "modmap", body{FQDN: fqdn, IP: ip, Checker: checker}, nil)
}

func (c *Client) DeleteFQDNMember(ctx context.Context, fqdn, ip string) error {
	type body struct {
		FQDN string `json:"fqdn"`
		IP   string `json:"ip"`
	}
	return c.call(ctx, "delmap", body{FQDN: fqdn, IP: ip}, nil)
}
