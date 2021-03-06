package chain

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/chain/chainpb"
)

type DetailsView struct {
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
	GoVersion  string `json:"go_version"`

	ClientInfo string `json:"client_info,omitempty"`

	ChainName string `json:"chain_name,omitempty"`
	ChainType string `json:"chain_type,omitempty"`

	NodeName         string            `json:"node_name,omitempty"`
	NodeVersion      string            `json:"node_version,omitempty"`
	NodeLocalPeerUID string            `json:"node_local_peer_uid,omitempty"`
	NodeHealth       string            `json:"node_health,omitempty"`
	NodeRoles        []string          `json:"node_roles,omitempty"`
	NodeProperties   map[string]string `json:"node_properties,omitempty"`

	GenesisHash string `json:"genesis_hash,omitempty"`

	IndexingStarted          bool       `json:"indexing_started"`
	LastIndexVersion         int64      `json:"last_index_version,omitempty"`
	LastIndexedHeight        int64      `json:"last_indexed_height,omitempty"`
	LastIndexedSession       int64      `json:"last_indexed_session,omitempty"`
	LastIndexedSessionHeight int64      `json:"last_indexed_session_height,omitempty"`
	LastIndexedEra           int64      `json:"last_indexed_era,omitempty"`
	LastIndexedEraHeight     int64      `json:"last_indexed_era_height,omitempty"`
	LastSpecVersion          string     `json:"chain_spec_version,omitempty"`
	LastIndexedTime          types.Time `json:"last_indexed_time,omitempty"`
	LastIndexedAt            types.Time `json:"last_indexed_at,omitempty"`
	Lag                      int64      `json:"indexing_lag,omitempty"`
}

func ToDetailsView(recentSyncable *model.Syncable, headResponse *chainpb.GetHeadResponse, statusResponse *chainpb.GetStatusResponse,
	lastSessionHeight int64, lastEraHeight int64) *DetailsView {
	view := &DetailsView{
		AppName:    config.AppName,
		AppVersion: config.AppVersion,
		GoVersion:  config.GoVersion,
	}

	if statusResponse != nil {
		view.ClientInfo = statusResponse.GetClientInfo()
		view.ChainName = statusResponse.GetChainName()
		view.ChainType = statusResponse.GetChainType()
		view.NodeName = statusResponse.GetNodeName()
		view.NodeVersion = statusResponse.GetNodeVersion()
		view.NodeLocalPeerUID = statusResponse.GetNodeLocalPeerUid()
		view.NodeHealth = statusResponse.GetNodeHealth()
		view.NodeRoles = statusResponse.GetNodeRoles()
		view.NodeProperties = statusResponse.GetNodeProperties()
		view.GenesisHash = statusResponse.GetGenesisHash()
	}

	view.IndexingStarted = recentSyncable != nil
	if view.IndexingStarted {
		view.LastIndexVersion = recentSyncable.IndexVersion
		view.LastIndexedHeight = recentSyncable.Height
		view.LastIndexedSession = recentSyncable.Session
		view.LastIndexedSessionHeight = lastSessionHeight
		view.LastIndexedEra = recentSyncable.Era
		view.LastIndexedEraHeight = lastEraHeight
		view.LastSpecVersion = recentSyncable.SpecVersion
		view.LastIndexedTime = recentSyncable.Time
		view.LastIndexedAt = recentSyncable.CreatedAt

		if headResponse != nil {
			view.Lag = headResponse.Height - recentSyncable.Height
		}
	}

	return view
}
