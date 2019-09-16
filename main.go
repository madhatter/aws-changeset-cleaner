package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var sess *session.Session
var cfSvc *cloudformation.CloudFormation

type ChangeSet struct {
	ChangeSetId     string    `json:"ChangeSetId"`
	ChangeSetName   string    `json:"ChangeSetName"`
	CreationTime    time.Time `json:"CreationTime"`
	Description     string    `json:"Description"`
	ExecutionStatus string    `json:"ExecutionStatus"`
	StackId         string    `json:"StackId"`
	StackName       string    `json:"StackName"`
	Status          string    `json:"Status"`
	StatusReason    string    `json:"StatusReason"`
}

type ChangeSets struct {
	Sets []ChangeSet
}

func (set *ChangeSet) Test() string {
	return "Test"
}

// initializes the client for cloudformation
func createClient(profile *string) {
	sess, _ = session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String("eu-central-1")},
		Profile: *profile,
	})
	cfSvc = cloudformation.New(sess)
}

// fetches all the stacks
func fetchChangeSets(cfSvc *cloudformation.CloudFormation, stackName *string) (ChangeSets, error) {
	lcsInput := cloudformation.ListChangeSetsInput{
		StackName: stackName,
	}

	ntoken := "1"
	var changeset ChangeSet
	var changesets ChangeSets

	for ntoken != "" {
		output, err := cfSvc.ListChangeSets(&lcsInput)

		if err != nil {
			fmt.Println("Error", err)
			return changesets, err
		} else {
			if output.NextToken != nil {
				ntoken = *output.NextToken
				lcsInput.NextToken = &ntoken
			} else {
				ntoken = ""
			}

			for _, v := range output.Summaries {
				changeset.ChangeSetId = *v.ChangeSetId
				changeset.ChangeSetName = *v.ChangeSetName
				changeset.CreationTime = *v.CreationTime
				changeset.Description = *v.Description
				changeset.ExecutionStatus = *v.ExecutionStatus
				changeset.StackId = *v.StackId
				changeset.StackName = *v.StackName
				changeset.Status = *v.Status
				changeset.StatusReason = *v.StatusReason

				changesets.Sets = append(changesets.Sets, changeset)
			}
		}
	}

	return changesets, nil
}

// deletes all the failed changeSets for a given stackName
func deleteChangeSets(cfSvc *cloudformation.CloudFormation, stackName string) {
	lcsInput := cloudformation.ListChangeSetsInput{
		StackName: aws.String(stackName),
	}

	ntoken := "1"

	for ntoken != "" {
		output, err := cfSvc.ListChangeSets(&lcsInput)

		if err != nil {
			fmt.Println("Error", err)
		} else {
			if output.NextToken != nil {
				ntoken = *output.NextToken
				lcsInput.NextToken = &ntoken
			} else {
				ntoken = ""
			}

			for i, _ := range output.Summaries {
				if *output.Summaries[i].Status == "FAILED" {
					csName := *output.Summaries[i].ChangeSetName
					fmt.Println(csName)

					dcsInput := cloudformation.DeleteChangeSetInput{
						ChangeSetName: &csName,
						StackName:     aws.String("opal-inventory-ecr-live"),
					}
					req, err := cfSvc.DeleteChangeSet(&dcsInput)

					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(req)
					}
				}
			}
		}
	}
}

// deletes changeset in a given collection but keeping those newer than the given time
func deleteChangeSetsTimeGap(cfSvc *cloudformation.CloudFormation, sets *ChangeSets, limit *time.Time) error {
	for k, _ := range sets.Sets {
		fmt.Println(sets.Sets[k].CreationTime)
	}

	fmt.Println(limit.Format("15:04:05"))
	return nil
}

// deletes changesets in a given collection but keeping an also given number
func deleteChangeSetsKeep(cfSvc *cloudformation.CloudFormation, sets *ChangeSets, keep *int) error {
	for index := 0; index < len(sets.Sets)-*keep; index++ {
		if sets.Sets[index].Status == "FAILED" {
			csName := sets.Sets[index].ChangeSetName
			stack := sets.Sets[index].StackName
			csTime := sets.Sets[index].CreationTime

			fmt.Printf("Deleting changeset %s (%s) on stack %s.\n", csName, csTime, stack)

			dcsInput := cloudformation.DeleteChangeSetInput{
				ChangeSetName: &csName,
				StackName:     &stack,
			}

			req, err := cfSvc.DeleteChangeSet(&dcsInput)

			if err != nil {
				fmt.Println(req)
				return err
			}
		}
	}

	return nil
}

// the main function
func main() {
	//dateForLimit := time.Now()
	keep := 10
	profile := aws.String("dv-live-developer")

	createClient(profile)

	sets, err := fetchChangeSets(cfSvc, aws.String("opal-inventory-ecr-live"))

	if err != nil {
		fmt.Println(err)
	} else {
		err := deleteChangeSetsKeep(cfSvc, &sets, &keep)

		if err != nil {
			fmt.Println(err)
		}
	}

	//fmt.Println(sets.Sets[1].ChangeSetId)
	//deleteChangeSets("dv-live-developer", "opal-inventory-ecr-live")
}
