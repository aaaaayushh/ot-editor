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
	// if both operations are from the same client, and op1 was generated before op2, return op1
	// because op1 happened before op2, hence not transformations are needed
	if op1.ID.ClientID == op2.ID.ClientID && op1.ID.SequenceNum < op2.ID.SequenceNum {
		return op1
	}

	// else transform the operations according to op2's type
	if op2.Type == operation.Insert {
		return TransformInsert(op1, op2)
	}
	return TransformDelete(op1, op2)
}

// TransformInsert transforms op1 according to op2, which is supposed to be an INSERTION
func TransformInsert(op1, op2 *operation.Operation) *operation.Operation {
	if op1.Type == operation.Insert {
		/* if op1 is also an insertion
		1. if op1's position is less than op2's position, no transformation is needed OR
		2. The rule used here is that if two insert operations have the same position, the operation from the client
			with the smaller ClientID is considered to have occurred first. This is an arbitrary rule, but it ensures
			consistency across all clients.
		*/
		if op1.Position < op2.Position || (op1.Position == op2.Position && op1.ID.ClientID < op2.ID.ClientID) {
			return op1
		}
		// otherwise shift the position of the insert operation by op2's content length
		op1.Position += len(op2.Content)
		return op1
	}
	// if op1 is a deletion, and op1's position is less than or equal to op2's position, no transformation is needed
	if op1.Position <= op2.Position {
		return op1
	}
	op1.Position += len(op2.Content)
	return op1
}

// TransformDelete transforms op1 according to op2, which is supposed to be a DELETION
func TransformDelete(op1, op2 *operation.Operation) *operation.Operation {
	if op1.Type == operation.Delete {
		/* if op1 is also a deletion
		`	1. if op1's position is less than op2's position, no transformation is needed OR
			2. The rule used here is that if two delete operations have the same position, the operation from the client
				with the smaller ClientID is considered to have occurred first. This is an arbitrary rule, but it ensures
				consistency across all clients.
		*/
		if op1.Position < op2.Position || (op1.Position == op2.Position && op1.ID.ClientID < op2.ID.ClientID) {
			return op1
		}

		// otherwise shift the position of the delete operation by op2's content length
		op1.Position -= len(op2.Content)
		return op1
	}
	// if op1 is an insertion, and op1's position is less than or equal to op2's position, no transformation is needed
	if op1.Position <= op2.Position {
		return op1
	}
	// if op1's position is less than op2's position + op2's content length, then op1's content is split into two parts
	// and the part that overlaps with op2's content is removed
	if op1.Position < op2.Position+len(op2.Content) {
		op1.Content = op1.Content[:op2.Position-op1.Position] + op1.Content[op2.Position-op1.Position+len(op2.Content):]
		op1.Position = op2.Position
	} else {
		op1.Position -= len(op2.Content)
	}
	return op1
}
