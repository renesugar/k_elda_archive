package command

import (
	"testing"

	clientMock "github.com/kelda/kelda/api/client/mocks"
	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
)

func TestStopNamespaceDefault(t *testing.T) {
	t.Parallel()

	c := new(clientMock.Client)
	c.On("QueryBlueprints").Once().Return([]db.Blueprint{{
		Blueprint: blueprint.Blueprint{Namespace: "testSpace",
			Machines: []blueprint.Machine{{}}}}}, nil)
	c.On("Deploy", mock.Anything).Return(nil)

	stopCmd := NewStopCommand()
	stopCmd.force = true
	stopCmd.client = c
	stopCmd.Run()

	c.AssertCalled(t, "Deploy", blueprint.Blueprint{Namespace: "testSpace"}.String())

	c.On("QueryBlueprints").Return(nil, nil)
	stopCmd.Run()
	c.AssertNumberOfCalls(t, "Deploy", 1)
}

func TestStopNamespace(t *testing.T) {
	t.Parallel()

	c := &clientMock.Client{}
	c.On("QueryBlueprints").Return(nil, nil)
	c.On("Deploy", mock.Anything).Return(nil)

	stopCmd := NewStopCommand()
	stopCmd.client = c
	stopCmd.force = true
	stopCmd.namespace = "namespace"
	stopCmd.Run()

	c.AssertCalled(t, "Deploy", blueprint.Blueprint{Namespace: "namespace"}.String())
}

func TestStopContainers(t *testing.T) {
	t.Parallel()

	c := &clientMock.Client{}
	c.On("QueryBlueprints").Return([]db.Blueprint{{
		Blueprint: blueprint.Blueprint{
			Namespace: "testSpace",
			Machines: []blueprint.Machine{
				{Provider: "Amazon"},
				{Provider: "Google"}},
			Containers: []blueprint.Container{{}, {}}},
	}}, nil)

	c.On("Deploy", mock.Anything).Return(nil)

	stopCmd := NewStopCommand()
	stopCmd.client = c
	stopCmd.onlyContainers = true
	stopCmd.force = true
	stopCmd.Run()

	c.AssertCalled(t, "Deploy", blueprint.Blueprint{
		Namespace: "testSpace",
		Machines: []blueprint.Machine{{
			Provider: "Amazon",
		}, {
			Provider: "Google",
		}}}.String())

}

func TestStopFlags(t *testing.T) {
	t.Parallel()

	expNamespace := "namespace"
	checkStopParsing(t, []string{"-namespace", expNamespace},
		Stop{namespace: expNamespace}, nil)
	checkStopParsing(t, []string{"-f"}, Stop{force: true}, nil)
	checkStopParsing(t, []string{"-f", expNamespace},
		Stop{force: true, namespace: expNamespace}, nil)
	checkStopParsing(t, []string{expNamespace}, Stop{namespace: expNamespace}, nil)
	checkStopParsing(t, []string{}, Stop{}, nil)
}

func checkStopParsing(t *testing.T, args []string, expFlags Stop, expErr error) {
	stopCmd := NewStopCommand()
	err := parseHelper(stopCmd, args)

	assert.Equal(t, expErr, err)
	assert.Equal(t, expFlags.namespace, stopCmd.namespace)
	assert.Equal(t, expFlags.force, stopCmd.force)
}

func TestStopPromptsUser(t *testing.T) {
	oldConfirm := confirm
	defer func() {
		confirm = oldConfirm
	}()

	compile = func(path string, args []string) (blueprint.Blueprint, error) {
		return blueprint.Blueprint{}, nil
	}

	for _, confirmResp := range []bool{true, false} {
		confirm = func(in io.Reader, prompt string) (bool, error) {
			return confirmResp, nil
		}

		c := new(clientMock.Client)
		c.On("QueryBlueprints").Return([]db.Blueprint{{
			Blueprint: blueprint.Blueprint{
				Namespace:  "ns",
				Machines:   []blueprint.Machine{{}, {}},
				Containers: []blueprint.Container{{}},
			},
		}}, nil)
		c.On("Deploy", blueprint.Blueprint{Namespace: "ns"}.String()).Return(nil)

		stopCmd := NewStopCommand()
		stopCmd.client = c
		stopCmd.Run()

		if confirmResp {
			c.AssertCalled(t, "Deploy", mock.Anything)
		} else {
			c.AssertNotCalled(t, "Deploy", mock.Anything)
		}
	}
}
