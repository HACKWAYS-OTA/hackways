ALTER TABLE events ADD COLUMN faq_text TEXT NOT NULL DEFAULT '' CHECK(char_length(faq_text)<=10000),
 ADD COLUMN contact_url TEXT NOT NULL DEFAULT '' CHECK(char_length(contact_url)<=2048),
 ADD COLUMN rewards TEXT NOT NULL DEFAULT '' CHECK(char_length(rewards)<=10000);
