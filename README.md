# aws-changeset-cleaner
A small tool to delete failed cloudformation changesets

## Purpose
When using the aws cli to deploy cloudformation templates without changes you generate changesets that are marked as failed.
It seems there is an undocumented hardlimit of 1000 (failed?) changesets. If you hit that you are not able to deploy to that stack anymore.

This cli tool deletes those changesets. There also is a flag for the paranoid to keep a number of those and just delete the rest.

## Example
```
./changeset-cleaner --profile <YOUR_AWS_PROFILE> \
  --stack <STACK_BLOCKED_BY_TOO_MANY_FAILED_CHANGESETS> \
  --keep 100
```
If no stack name is given all stacks will be cleaned from failed changesets.
