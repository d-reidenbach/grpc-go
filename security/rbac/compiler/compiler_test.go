package compiler

import (
	"fmt"
	"testing"
)

// type request struct {
// 	url_path string
// }
// type connection struct {
// 	uri_san_peer_certificate string
// }

var testAction = "ALLOW"
var testPolicies = make(map[string]string)

func TestApp(t *testing.T) {
	testPolicies["test access"] = "request.url_path.startsWith('/pkg.service/test')"
	testPolicies["admin access"] = "connection.uri_san_peer_certificate == 'cluster/ns/default/sa/admin'"
	testPolicies["dev access"] = "request.url_path == '/pkg.service/dev' && connection.uri_san_peer_certificate == 'cluster/ns/default/sa/dev'"
	env := createUserPolicyCelEnv()
	rbac := compileYamltoRbac("user_policy.yaml")
	// fmt.Println(rbac.String())
	policies := rbac.Policies

	for name, rbacPolicy := range policies {
		testPolicy := testPolicies[name]
		fmt.Println(testPolicy)
		testAst := compileCel(env, testPolicy)
		testProgram, _ := env.Program(testAst)

		expr := rbacPolicy.Condition
		program := exprToProgram(env, expr)

		vars := map[string]interface{}{
			"request.url_path":                    "/pkg.service/test",
			"connection.uri_san_peer_certificate": "cluster/ns/default/sa/admin",
		}

		got, _, gotErr := (*program).Eval(vars)
		if gotErr != nil {
			t.Errorf("Error in evaluating CEL program %s", gotErr.Error())
		}
		want, _, wantErr := testProgram.Eval(vars)
		if wantErr != nil {
			t.Errorf("Error in evaluating TEST CEL program %s", wantErr.Error())
		}

		if got != want {
			t.Errorf("Error CEL prgram evaluations do not amtch up %v, %v", got, want)
		}
		fmt.Printf("Compiled rbac evaluation result: %v, Direct Evaluation result: %v \n", got, want)

		// get evaluatable program from rbacPolicy and thedn evaluate both
		// need to create the test IMPORT NEEDED THINGS TO CREATE THE REQUEST OBJECT
	}

}
