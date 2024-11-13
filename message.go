package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// PostMessageParams are the parameters required to post a message to Slack.
type PostMessageParams struct {
	// Context which will be applied to the message.
	Context map[string]string

	// Details.
	Description string

	// Actions.
	Dashboard     string
	Documentation string

	// Icon which will be applied to this message.
	Icon string
}

// Validate the parameters.
func (p PostMessageParams) Validate() error {
	var errs []error

	if p.Description == "" {
		errs = append(errs, fmt.Errorf("description is required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// PostMessage to Slack channel.
func (c *Client) PostMessage(params PostMessageParams) error {
	if err := params.Validate(); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	var message Message

	if len(params.Context) != 0 {
		// Context which allows the developer to understand what project/environment is affected.
		// This is intentionally ordered as Environment/Project/Cluster because as a developer, I would expect someone
		// to review the message in that order.
		//   * Is it production?
		//   * Which project?
		//   * Which cluster?
		context := BlockContext{
			Type: BlockTypeContext,
		}

		for key, value := range params.Context {
			context.Elements = append(context.Elements, BlockContextElement{
				Type: BlockElementTypeMarkdown,
				Text: fmt.Sprintf("*%s* = %s", key, value),
			})
		}

		message.Blocks = append(message.Blocks, context)

		// Separate the context from the content.
		message.Blocks = append(message.Blocks, BlockDivider{
			Type: BlockTypeDivider,
		})
	}

	// Details of the alarm.
	details := BlockSection{
		Type: BlockTypeSection,
		Text: BlockSectionText{
			Type: BlockTextTypeMarkdown,
			Text: params.Description,
		},
	}

	if params.Icon != "" {
		details.Accessory = &BlockSectionAccessory{
			Type:     BlockElementTypeImage,
			ImageURL: params.Icon,
			AltText:  "Identifier for the Slack message",
		}
	}

	message.Blocks = append(message.Blocks, details)

	// Links which can be used to action message eg. Go to this dashboard or this documentation.
	var links []string

	if params.Dashboard != "" {
		links = append(links, fmt.Sprintf("<%s|:skpr_dashboard: Review with Dashboard>", params.Dashboard))
	}

	if params.Documentation != "" {
		links = append(links, fmt.Sprintf("<%s|:skpr_documentation: Triage using Documentation>", params.Documentation))
	}

	if len(links) > 0 {
		message.Blocks = append(message.Blocks, BlockDivider{
			Type: BlockTypeDivider,
		})

		message.Blocks = append(message.Blocks, BlockSection{
			Type: BlockTypeSection,
			Text: BlockSectionText{
				Type: BlockTextTypeMarkdown,
				Text: "*Next Steps:*",
			},
		})

		message.Blocks = append(message.Blocks, BlockSection{
			Type: BlockTypeSection,
			Text: BlockSectionText{
				Type: BlockTextTypeMarkdown,
				Text: strings.Join(links, "\t"),
			},
		})
	}

	request, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for _, webhook := range c.webhooks {
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(request))
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)

		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("returned status code: %d", resp.StatusCode)
		}
	}

	return nil
}
