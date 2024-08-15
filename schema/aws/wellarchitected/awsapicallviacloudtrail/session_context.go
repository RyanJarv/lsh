package awsapicallviacloudtrail

type SessionContext struct {
    Attributes Attributes `json:"attributes"`
    SessionIssuer SessionIssuer `json:"sessionIssuer"`
    WebIdFederationData interface{} `json:"webIdFederationData,omitempty"`
}

func (s *SessionContext) SetAttributes(attributes Attributes) {
    s.Attributes = attributes
}

func (s *SessionContext) SetSessionIssuer(sessionIssuer SessionIssuer) {
    s.SessionIssuer = sessionIssuer
}

func (s *SessionContext) SetWebIdFederationData(webIdFederationData interface{}) {
    s.WebIdFederationData = webIdFederationData
}