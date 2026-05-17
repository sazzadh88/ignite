// Package mail provides email sending functionality with support for multiple transports.
package mail

import (
	"fmt"
	"mime"
	"path/filepath"
)

// Address represents an email address with an optional display name.
type Address struct {
	Name    string
	Address string
}

// String formats the address for email headers.
func (a Address) String() string {
	if a.Name != "" {
		return fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", a.Name), a.Address)
	}
	return a.Address
}

// Attachment represents an email attachment.
type Attachment struct {
	Name     string
	Content  []byte
	MimeType string
}

// Message represents an email message.
type Message struct {
	From        []Address
	To          []Address
	Cc          []Address
	Bcc         []Address
	ReplyTo     []Address
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
	Headers     map[string]string
	Tags        []string
	Metadata    map[string]string
}

// NewMessage creates a new message instance.
func NewMessage() *Message {
	return &Message{
		From:        []Address{},
		To:          []Address{},
		Cc:          []Address{},
		Bcc:         []Address{},
		ReplyTo:     []Address{},
		Attachments: []Attachment{},
		Headers:     make(map[string]string),
		Tags:        []string{},
		Metadata:    make(map[string]string),
	}
}

// SetFrom sets the from address.
func (m *Message) SetFrom(address string, name ...string) *Message {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	m.From = []Address{addr}
	return m
}

// AddTo adds a recipient.
func (m *Message) AddTo(address string, name ...string) *Message {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	m.To = append(m.To, addr)
	return m
}

// AddCc adds a CC recipient.
func (m *Message) AddCc(address string, name ...string) *Message {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	m.Cc = append(m.Cc, addr)
	return m
}

// AddBcc adds a BCC recipient.
func (m *Message) AddBcc(address string, name ...string) *Message {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	m.Bcc = append(m.Bcc, addr)
	return m
}

// SetReplyTo sets the reply-to address.
func (m *Message) SetReplyTo(address string, name ...string) *Message {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	m.ReplyTo = []Address{addr}
	return m
}

// SetSubject sets the subject.
func (m *Message) SetSubject(subject string) *Message {
	m.Subject = subject
	return m
}

// SetBody sets the plain text body.
func (m *Message) SetBody(body string) *Message {
	m.Body = body
	return m
}

// SetHTML sets the HTML body.
func (m *Message) SetHTML(html string) *Message {
	m.HTMLBody = html
	return m
}

// Attach attaches a file with the given content.
func (m *Message) Attach(name string, content []byte, mimeType ...string) *Message {
	mt := "application/octet-stream"
	if len(mimeType) > 0 && mimeType[0] != "" {
		mt = mimeType[0]
	} else {
		ext := filepath.Ext(name)
		if ext != "" {
			mt = mime.TypeByExtension(ext)
			if mt == "" {
				mt = "application/octet-stream"
			}
		}
	}
	m.Attachments = append(m.Attachments, Attachment{
		Name:     name,
		Content:  content,
		MimeType: mt,
	})
	return m
}

// AttachData attaches raw data with the given filename.
func (m *Message) AttachData(name string, data []byte) *Message {
	return m.Attach(name, data)
}

// SetHeader sets a custom header.
func (m *Message) SetHeader(key, value string) *Message {
	m.Headers[key] = value
	return m
}

// AddTag adds a tag to the message.
func (m *Message) AddTag(tag string) *Message {
	m.Tags = append(m.Tags, tag)
	return m
}

// SetMetadata sets metadata key-value pair.
func (m *Message) SetMetadata(key, value string) *Message {
	m.Metadata[key] = value
	return m
}
