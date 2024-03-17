package ot

import (
	"github.com/aaaaayushh/ot_editor/server/client"
	"github.com/aaaaayushh/ot_editor/server/operation"
)

func TransformOperation(op *operation.Operation, c *client.Client) *operation.Operation {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	for _, pastOp := range c.HistoryBuffer {
		op = TransformPair(op, &pastOp)
	}
	return op
}

func TransformPair(op1, op2 *operation.Operation) *operation.Operation {
	if op1.ID.ClientID == op2.ID.ClientID && op1.ID.SequenceNum < op2.ID.SequenceNum {
		return op1
	}

	if op2.Type == operation.Insert {
		return TransformInsert(op1, op2)
	}
	return TransformDelete(op1, op2)
}

func TransformInsert(op1, op2 *operation.Operation) *operation.Operation {
	if op1.Type == operation.Insert {
		if op1.Position < op2.Position || (op1.Position == op2.Position && op1.ID.ClientID < op2.ID.ClientID) {
			return op1
		}
		op1.Position += len(op2.Content)
		return op1
	}
	if op1.Position <= op2.Position {
		return op1
	}
	op1.Position += len(op2.Content)
	return op1
}

func TransformDelete(op1, op2 *operation.Operation) *operation.Operation {
	if op1.Type == operation.Delete {
		if op1.Position < op2.Position || (op1.Position == op2.Position && op1.ID.ClientID < op2.ID.ClientID) {
			return op1
		}
		op1.Position -= len(op2.Content)
		return op1
	}
	if op1.Position <= op2.Position {
		return op1
	}
	if op1.Position < op2.Position+len(op2.Content) {
		op1.Content = op1.Content[:op2.Position-op1.Position] + op1.Content[op2.Position-op1.Position+len(op2.Content):]
		op1.Position = op2.Position
	} else {
		op1.Position -= len(op2.Content)
	}
	return op1
}
