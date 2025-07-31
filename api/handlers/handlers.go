package handlers

import (
	"net/http"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo/v4"
)

type OrgSetup struct {
	OrgName      string
	MSPID        string
	CryptoPath   string
	CertPath     string
	KeyPath      string
	TLSCertPath  string
	PeerEndpoint string
	GatewayPeer  string
	Gateway      client.Gateway
}

// QueryPayload defines the payload required to query the chaincode.
// The fields need to be exported so that Echo can bind incoming JSON
// data to this struct.
type QueryPayload struct {
	ChaincodeID string   `json:"chaincodeID"`
	ChannelID   string   `json:"channelID"`
	Function    string   `json:"function"`
	Args        []string `json:"args"`
}

// InvokePayload defines the payload required to invoke a transaction on the
// chaincode. Fields are exported for JSON binding.
type InvokePayload struct {
	ChaincodeID string   `json:"chaincodeID"`
	ChannelID   string   `json:"channelID"`
	Function    string   `json:"function"`
	Args        []string `json:"args"`
}

func BaseRoute(e echo.Context) error {
	return e.JSON(http.StatusOK, echo.Map{
		"message": "Server is Healthy.",
	})
}

func (setup OrgSetup) Query(e echo.Context) error {

	queryPayload := new(QueryPayload)
	if err := e.Bind(queryPayload); err != nil {
		return e.JSON(http.StatusBadRequest, echo.Map{
			"message": "Unable to create Query.",
		})
	}

	network := setup.Gateway.GetNetwork(queryPayload.ChannelID)
	contract := network.GetContract(queryPayload.ChaincodeID)
	evalResponse, err := contract.EvaluateTransaction(queryPayload.Function, queryPayload.Args...)
	if err != nil {
		return e.JSON(http.StatusBadRequest, echo.Map{
			"message": "Unable to evaluate Response.",
		})
	}

	return e.JSON(http.StatusOK, echo.Map{
		"data": evalResponse,
	})
}

func (setup *OrgSetup) Invoke(e echo.Context) error {
	invokePayload := new(InvokePayload)
	if err := e.Bind(invokePayload); err != nil {
		return e.JSON(http.StatusBadRequest, echo.Map{
			"message": "Unable to create Query.",
		})
	}

	network := setup.Gateway.GetNetwork(invokePayload.ChannelID)
	contract := network.GetContract(invokePayload.ChaincodeID)
	txnProposal, err := contract.NewProposal(invokePayload.Function,
		client.WithArguments(invokePayload.Args...))
	if err != nil {
		return e.JSON(http.StatusBadRequest, echo.Map{
			"message": "Unable to create Txn proposal.",
		})
	}

	txnEndorsed, err := txnProposal.Endorse()
	if err != nil {
		return e.JSON(http.StatusBadRequest, echo.Map{
			"message": "Unable endorse Txn proposal.",
		})
	}
	txnCommitted, err := txnEndorsed.Submit()
	if err != nil {
		return e.JSON(http.StatusBadRequest, echo.Map{
			"message": "Unable commit endorsed Txn.",
		})
	}

	return e.JSON(http.StatusOK, echo.Map{
		"data": txnCommitted,
	})
}
