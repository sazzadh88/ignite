package notifications

// SlackAttachment represents a Slack message attachment.
type SlackAttachment struct {
	Color      string
	Title      string
	Text       string
	Footer     string
	FooterIcon string
	Fields     []SlackField
}

// SlackField represents a field in a Slack attachment.
type SlackField struct {
	Title string
	Value string
	Short bool
}

// SlackMessage represents a Slack notification message.
type SlackMessage struct {
	Content     string
	Attachments []SlackAttachment
}

// NewSlackMessage creates a new Slack message.
func NewSlackMessage() *SlackMessage {
	return &SlackMessage{
		Attachments: make([]SlackAttachment, 0),
	}
}

// SetContent sets the main content of the Slack message.
func (s *SlackMessage) SetContent(content string) *SlackMessage {
	s.Content = content
	return s
}

// AddAttachment adds an attachment to the message using a builder function.
func (s *SlackMessage) AddAttachment(fn func(*SlackAttachment)) *SlackMessage {
	att := &SlackAttachment{
		Fields: make([]SlackField, 0),
	}
	fn(att)
	s.Attachments = append(s.Attachments, *att)
	return s
}

// SetColor sets the color of the attachment.
func (a *SlackAttachment) SetColor(color string) *SlackAttachment {
	a.Color = color
	return a
}

// SetTitle sets the title of the attachment.
func (a *SlackAttachment) SetTitle(title string) *SlackAttachment {
	a.Title = title
	return a
}

// SetText sets the text of the attachment.
func (a *SlackAttachment) SetText(text string) *SlackAttachment {
	a.Text = text
	return a
}

// SetFooter sets the footer of the attachment.
func (a *SlackAttachment) SetFooter(footer string) *SlackAttachment {
	a.Footer = footer
	return a
}

// AddField adds a field to the attachment.
func (a *SlackAttachment) AddField(title, value string, short bool) *SlackAttachment {
	a.Fields = append(a.Fields, SlackField{
		Title: title,
		Value: value,
		Short: short,
	})
	return a
}
