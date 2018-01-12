package asttest

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/ast"
)

var IgnoreBaseNodeOptions = []cmp.Option{
	cmpopts.IgnoreFields(ast.ArrayExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.ArrowFunctionExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.BinaryExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.BlockStatement{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.BooleanLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.CallExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.ConditionalExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.DateTimeLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.DurationLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.ExpressionStatement{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.FloatLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.Identifier{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.ImportDeclaration{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.IntegerLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.LogicalExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.MemberExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.ObjectExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.File{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.Property{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.RegexpLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.ReturnStatement{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.StringLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.UnaryExpression{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.UnsignedIntegerLiteral{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.VariableDeclaration{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.VariableDeclarator{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.VersionDeclaration{}, "BaseNode"),
	cmpopts.IgnoreFields(ast.VersionNumber{}, "BaseNode"),
}
