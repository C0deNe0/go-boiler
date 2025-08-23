package email

func (c *Client) SendWelcomeEmail(to, firstName string) error {
	data := map[string]string{

		"UserFirstName": firstName,
	}

	return c.SendEmail(
		to,
		"welcome to BoilerPlate!",
		TemplateWelcome,
		data,
	)
}
