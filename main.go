package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

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

// fetches all the stacks
func fetchChangeSets(profile *string, stackName *string) ChangeSets {
	sess, _ := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String("eu-central-1")},
		Profile: *profile,
	})

	cfSvc := cloudformation.New(sess)
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
	fmt.Println(changesets.Sets[0].ChangeSetName)

	return changesets
}

// deletes all the changeSets for a given stackName
func deleteChangeSets(profile string, stackName string) {
	sess, _ := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String("eu-central-1")},
		Profile: profile,
	})

	cfSvc := cloudformation.New(sess)
	lcsInput := cloudformation.ListChangeSetsInput{
		StackName: aws.String("opal-inventory-ecr-live"),
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

func deleteChangeSetsTimeGap(profile *string, sets *ChangeSets, limit *time.Time) error {
	for k, _ := range sets.Sets {
		fmt.Println(sets.Sets[k].CreationTime)
	}

	fmt.Println(limit.Format("15:04:05"))
	return nil
}

func deleteChangeSetsKeep(profile *string, sets *ChangeSets, keep *int) error {
	for index := 0; index < len(sets.Sets)-*keep; index++ {
		fmt.Println(sets.Sets[index].CreationTime)
	}

	return nil
}

func main() {
	//dateForLimit := time.Now()
	keep := 10

	sets := fetchChangeSets(aws.String("dv-live-developer"), aws.String("opal-inventory-ecr-live"))
	//err := deleteChangeSetsTimeGap(aws.String("dv-live-developer"), &sets, &dateForLimit)
	err := deleteChangeSetsKeep(aws.String("dv-live-developer"), &sets, &keep)

	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println(sets.Sets[1].ChangeSetId)
	//deleteChangeSets("dv-live-developer", "opal-inventory-ecr-live")
}
