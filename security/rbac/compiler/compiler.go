package compiler

import (
	"fmt"
	"io/ioutil"
	"log"

	pb "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2" // used to be v2
	"github.com/golang/glog"
	cel "github.com/google/cel-go/cel"
	decls "github.com/google/cel-go/checker/decls"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// UserPolicy is the user policy
type UserPolicy struct {
	Action string `yaml:"action"`
	Rules  []struct {
		Name      string `yaml:"name"`
		Condition string `yaml:"condition"`
	} `yaml:"rules"`
}

func readYaml(filePath string) []byte {
	// IS THIS REALLY A FILE PATH OR JUST FILE NAME. WHAT HAPPENS WHEN THE FILE IS NOT IN IMMEDIETE DIRECTORY
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error in reading yaml: %v", err)
	}
	fmt.Println("Read File Success")
	return yamlFile
}

func parseYaml(file []byte, policy *UserPolicy) {
	err := yaml.Unmarshal(file, policy)
	if err != nil {
		log.Fatalf("Failed in parsing of yaml")
	}
	fmt.Println("Parse File Success")

}

func createUserPolicyCelEnv() *cel.Env {
	env, _ := cel.NewEnv(
		cel.Declarations(
			decls.NewIdent("request.url_path", decls.String, nil),
			decls.NewIdent("request.host", decls.String, nil),
			decls.NewIdent("request.method", decls.String, nil),
			decls.NewIdent("request.headers", decls.NewMapType(decls.String, decls.String), nil),
			decls.NewIdent("source.address", decls.String, nil),
			decls.NewIdent("source.port", decls.Int, nil),
			decls.NewIdent("destination.address", decls.String, nil),
			decls.NewIdent("destination.port", decls.Int, nil),
			decls.NewIdent("connection.uri_san_peer_certificate", decls.String, nil)))
	// ast := compile(env, `source.address.startsWith('1.1.1.') || request.headers['type'] == 'foo'`, decls.Bool)
	// checked, iss := env.Check(ast)
	// if iss != nil && iss.Err() != nil {
	// 		glog.Exit(iss.Err())
	// }
	// program, _ := env.Program(checked)
	return env
}

func compileCel(env *cel.Env, condition string) *cel.Ast {
	ast, iss := env.Parse(condition)
	// Report syntactic errors, if present.
	if iss.Err() != nil {
		glog.Exit(iss.Err())
	}
	// Type-check the expression for correctness.
	checked, iss := env.Check(ast)
	if iss.Err() != nil {
		glog.Exit(iss.Err())
	}
	// Check the result type is a string.
	if !proto.Equal(checked.ResultType(), decls.Bool) {
		glog.Exitf(
			"Got %v, wanted %v result type",
			checked.ResultType(), decls.String)
	}

	return checked
}
func astToCheckedExpr(checked *cel.Ast) *expr.CheckedExpr {
	checkedExpr, err := cel.AstToCheckedExpr(checked) // v3
	if err != nil {
		log.Fatalf("Failed Converting AST to Checked Express %v", err)
	}
	// checkedExpr := checked.Expr() // v2
	return checkedExpr
}
func compileYamltoRbac(filename string) pb.RBAC {
	// "user_policy.yaml"
	yamlFile := readYaml(filename)
	var userPolicy UserPolicy
	parseYaml(yamlFile, &userPolicy)
	fmt.Println(userPolicy)
	fmt.Println("____________________________________")
	fmt.Println(" ")

	env := createUserPolicyCelEnv()

	fmt.Println("Finished CEL Environment starting RBAC Loop")
	// rules := policy.Rules
	var rbac pb.RBAC
	rbac.Action = pb.RBAC_Action(pb.RBAC_Action_value[userPolicy.Action])
	rbac.Policies = make(map[string]*pb.Policy)

	for index := range userPolicy.Rules {
		rule := userPolicy.Rules[index]
		name := rule.Name
		condition := rule.Condition
		var policy pb.Policy
		// Check that the expression compiles and returns a String.
		checked := compileCel(env, condition)

		// checkedExpr := astToCheckedExpr(checked) // v3
		// policy.CheckedCondition = checkedExpr // v3

		checkedExpr := checked.Expr()  // v2
		policy.Condition = checkedExpr // v2

		rbac.Policies[name] = &policy

	}

	return rbac
}

// Converts an expression to a parsed expression, with SourceInfo nil.
func exprToParsedExpr(condition *expr.Expr) *expr.ParsedExpr {
	return &expr.ParsedExpr{Expr: condition}
}

// Converts an expression to a CEL program.
func exprToProgram(env *cel.Env, condition *expr.Expr) *cel.Program {
	// v3: can replace line with ast := cel.CheckedExprToAst(checkedExpr)
	ast := cel.ParsedExprToAst(exprToParsedExpr(condition))
	program, _ := env.Program(ast)
	return &program
}

func compile(inputFilename string, outputFilename string) {
	// rbac := compileYamltoRbac(inputFilename)
	// serialized, err := proto.Marshal(pb.ProtoReflect(rbac))
	// if err != nil {
	// 	log.Fatalf("Failed to Serialize RBAC Proto %v", err)
	// }
	// ioutil.WriteFile(outputFilename, serialized, 0644)

}
