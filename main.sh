#!/usr/bin/env bash

export GITHUB_BRANCH=${GITHUB_REF##*heads/}
export CI_SCRIPT_OPTIONS="ci_script_options"

COMMIT_MESSAGE=$(jq -r '.commits[-1].message' < "$GITHUB_EVENT_PATH")

hosts_file="$GITHUB_WORKSPACE/.github/hosts.yml"

if [[ -z "$SLACK_CHANNEL" ]]; then
	if [[ -f "$hosts_file" ]]; then
		user_slack_channel=$(shyaml get-value "$CI_SCRIPT_OPTIONS.slack-channel" < "$hosts_file" | tr '[:upper:]' '[:lower:]')
	fi
fi

if [[ -n "$user_slack_channel" ]]; then
	SLACK_CHANNEL="$user_slack_channel"
fi

# Check vault only if SLACK_WEBHOOK is empty.
if [[ -z "$SLACK_WEBHOOK" ]]; then

	# Login to vault using GH Token
	if [[ -n "$VAULT_GITHUB_TOKEN" ]]; then
		unset VAULT_TOKEN
		vault login -method=github token="$VAULT_GITHUB_TOKEN" > /dev/null
	fi

	if [[ -n "$VAULT_GITHUB_TOKEN" ]] || [[ -n "$VAULT_TOKEN" ]]; then
		SLACK_WEBHOOK=$(vault read -field=webhook secret/slack)
	fi
fi

if [[ -z "$SLACK_MESSAGE" ]]; then
	SLACK_MESSAGE="$COMMIT_MESSAGE"
fi

export SLACK_CHANNEL
export SLACK_MESSAGE
export SLACK_WEBHOOK

slack-notify "$@"
