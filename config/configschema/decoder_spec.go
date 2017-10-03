package configschema

import (
	"github.com/hashicorp/hcl2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// DecoderSpec returns a zcldec.Spec that can be used to decode a zcl Body
// using the facilities in the zcldec package.
//
// The returned specification is guaranteed to return a value of the same type
// returned by method ImpliedType, but it may contain null or unknown values if
// any of the block attributes are defined as optional and/or computed
// respectively.
func (b *Block) DecoderSpec() hcldec.Spec {
	ret := hcldec.ObjectSpec{}
	if b == nil {
		return ret
	}

	// If the behavior here is changed, usually the behavior of ImpliedType
	// must be changed to match. It is required that this method produce
	// a specification that decodes into a value of the implied type.

	for name, attrS := range b.Attributes {
		if attrS.Computed {
			ret[name] = &hcldec.LiteralSpec{
				Value: cty.UnknownVal(attrS.Type),
			}
		} else {
			ret[name] = &hcldec.AttrSpec{
				Type:     attrS.Type,
				Required: attrS.Required,
			}
		}
	}

	for name, blockS := range b.BlockTypes {
		if _, exists := ret[name]; exists {
			// This indicates an invalid schema, since it's not valid to
			// define both an attribute and a block type of the same name.
			// However, we don't raise this here since it's checked by
			// InternalValidate.
			continue
		}

		childSpec := blockS.Block.DecoderSpec()

		switch blockS.Nesting {
		case NestingSingle:
			ret[name] = &hcldec.BlockSpec{
				TypeName: name,
				Nested:   childSpec,
				Required: blockS.MinItems == 1 && blockS.MaxItems >= 1,
			}
		case NestingList:
			ret[name] = &hcldec.BlockListSpec{
				TypeName: name,
				Nested:   childSpec,
				MinItems: blockS.MinItems,
				MaxItems: blockS.MaxItems,
			}
		case NestingSet:
			ret[name] = &hcldec.BlockSetSpec{
				TypeName: name,
				Nested:   childSpec,
				MinItems: blockS.MinItems,
				MaxItems: blockS.MaxItems,
			}
		case NestingMap:
			ret[name] = &hcldec.BlockMapSpec{
				TypeName: name,
				Nested:   childSpec,
			}
		default:
			// Invalid nesting type is just ignored. It's checked by
			// InternalValidate.
			continue
		}
	}

	return ret
}
