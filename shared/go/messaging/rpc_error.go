package messaging

// rpcErrorPayload distinguishes a failed RPC reply from successful DTOs that
// may also contain an `error` field.
type rpcErrorPayload struct {
	Error    string `json:"error"`
	RPCError bool   `json:"_rpc_error"`
}
