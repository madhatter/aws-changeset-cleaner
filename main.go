package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var profile string
var processall = true
var stackToClean string
var keep = 10
var verbose = false
var sess *session.Session
var cfSvc *cloudformation.CloudFormation

// ChangeSet is public struct to store changesets
type ChangeSet struct {
	changeSetID     string
	changeSetName   string
	creationTime    time.Time
	description     string
	executionStatus string
	stackID         string
	stackName       string
	status          string
	statusReason    string
}

// ChangeSets is public collection of ChangeSets
type ChangeSets struct {
	sets []ChangeSet
}

// Stack to store information about the cloudformation stack
type Stack struct {
	creationTime        time.Time
	deletionTime        time.Time
	driftInformation    cloudformation.StackDriftInformationSummary
	lastUpdatedTime     time.Time
	stackID             string
	stackName           string
	stackStatus         string
	templateDescription string
}

// Stacks is public collection of Stacks
type Stacks struct {
	stacks []Stack
}

// initializes the client for cloudformation
func createClient(profile *string, verbose bool) {
	config := aws.Config{Region: aws.String("eu-central-1"), MaxRetries: aws.Int(15)}
	if verbose == true {
		config.WithLogLevel(aws.LogDebugWithRequestRetries)
	}

	sess, _ = session.NewSessionWithOptions(session.Options{
		Config:  config,
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
			fmt.Println("Error: ", err)
			return changesets, err
		}

		if output.NextToken != nil {
			ntoken = *output.NextToken
			lcsInput.NextToken = &ntoken
		} else {
			ntoken = ""
		}

		for _, v := range output.Summaries {
			changeset.changeSetID = *v.ChangeSetId
			changeset.changeSetName = *v.ChangeSetName
			changeset.creationTime = *v.CreationTime
			changeset.description = *v.Description
			changeset.executionStatus = *v.ExecutionStatus
			changeset.stackID = *v.StackId
			changeset.stackName = *v.StackName
			changeset.status = *v.Status
			changeset.statusReason = *v.StatusReason

			changesets.sets = append(changesets.sets, changeset)
		}
	}

	fmt.Printf("%s: %v failed changesets found.\n", changesets.sets[len(changesets.sets)-1].stackName, len(changesets.sets))
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

			for i := range output.Summaries {
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
	for k := range sets.sets {
		fmt.Println(sets.sets[k].creationTime)
	}

	fmt.Println(limit.Format("15:04:05"))
	return nil
}

// deletes changesets in a given collection but keeping an also given number
func deleteChangeSetsKeep(cfSvc *cloudformation.CloudFormation, sets *ChangeSets, keep *int) {
	for index := 0; index < len(sets.sets)-*keep; index++ {
		if sets.sets[index].status == "FAILED" {
			csName := sets.sets[index].changeSetName
			stack := sets.sets[index].stackName
			csTime := sets.sets[index].creationTime

			fmt.Printf("Deleting changeset %s (%s) on stack %s.\n", csName, csTime, stack)

			dcsInput := cloudformation.DeleteChangeSetInput{
				ChangeSetName: &csName,
				StackName:     &stack,
			}

			_, err := cfSvc.DeleteChangeSet(&dcsInput)

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func fetchStacks(cfSvc *cloudformation.CloudFormation) (Stacks, error) {
	lsInput := cloudformation.ListStacksInput{}

	ntoken := "1"

	var stack Stack
	var stacks Stacks

	for ntoken != "" {
		output, err := cfSvc.ListStacks(&lsInput)

		if err != nil {
			fmt.Println("Error", err)
			return stacks, err
		}
		if output.NextToken != nil {
			ntoken = *output.NextToken
			lsInput.NextToken = &ntoken
		} else {
			ntoken = ""
		}

		for _, v := range output.StackSummaries {
			//fmt.Println(*v)
			stack.creationTime = *v.CreationTime
			if *v.StackStatus == "DELETE_COMPLETE" {
				stack.deletionTime = *v.DeletionTime
			}
			stack.driftInformation = *v.DriftInformation
			if *v.StackStatus == "UPDATE_COMPLETE" {
				stack.lastUpdatedTime = *v.LastUpdatedTime
			}
			stack.stackID = *v.StackId
			stack.stackName = *v.StackName
			stack.stackStatus = *v.StackStatus
			stacks.stacks = append(stacks.stacks, stack)
		}
	}

	return stacks, nil
}

func parseCLIArguments() {
	profilePtr := flag.String("profile", "", "AWS profile to use. (Required)")
	stackPtr := flag.String("stack", "all", "Stack to clean {all stacks|<stackname>}.")
	keepPtr := flag.Int("keep", 10, "Number of changesets to keep.")
	verbosePtr := flag.Bool("verbose", false, "Verbose logging.")
	flag.Parse()

	if *profilePtr == "" {
		fmt.Println("No profile set.")
		flag.PrintDefaults()
		os.Exit(3)
	} else {
		profile = *profilePtr
	}

	if *stackPtr == "all" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println()
		fmt.Print("Processing on all stacks. Deleting all failed changesets on _all_ stacks. Continue (y/n)? ")
		text, _ := reader.ReadString('\n')
		if text != "y\n" {
			fmt.Println("Coward.")
			os.Exit(3)
		}
	} else {
		processall = false
		stackToClean = *stackPtr
	}

	keep = *keepPtr
	verbose = *verbosePtr
}

func cleanUpAllStacks(keep *int) error {
	stacks, err := fetchStacks(cfSvc)

	if err != nil {
		return err
	}

	for _, v := range stacks.stacks {
		if v.stackStatus != "DELETE_COMPLETE" {
			sets, err := fetchChangeSets(cfSvc, aws.String(v.stackName))
			if err != nil {
				return err
			}
			go deleteChangeSetsKeep(cfSvc, &sets, keep)
		}
	}

	return nil
}

// the main function
func main() {
	parseCLIArguments()

	createClient(&profile, verbose)

	if processall == true {
		err := cleanUpAllStacks(&keep)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		sets, err := fetchChangeSets(cfSvc, &stackToClean)
		if err != nil {
			fmt.Println(err)
		} else {
			deleteChangeSetsKeep(cfSvc, &sets, &keep)
		}
	}
}
