ALTER TABLE event_registration_questions DROP CONSTRAINT event_registration_questions_kind_check;
ALTER TABLE event_registration_questions ADD CONSTRAINT event_registration_questions_kind_check CHECK(kind IN ('text','textarea','url','number','checkbox','select','phone','email','date','time','multiselect','section'));
ALTER TABLE event_registration_questions DROP CONSTRAINT event_registration_questions_position_check;
ALTER TABLE event_registration_questions ADD CONSTRAINT event_registration_questions_position_check CHECK(position BETWEEN 0 AND 49);
ALTER TABLE event_registration_questions ADD COLUMN IF NOT EXISTS placeholder TEXT NOT NULL DEFAULT '' CHECK(char_length(placeholder)<=180),
 ADD COLUMN IF NOT EXISTS validation JSONB NOT NULL DEFAULT '{}' CHECK(jsonb_typeof(validation)='object');
ALTER TABLE events ADD COLUMN IF NOT EXISTS registration_form_settings JSONB NOT NULL DEFAULT '{}' CHECK(jsonb_typeof(registration_form_settings)='object');
ALTER TABLE registration_submissions DROP CONSTRAINT registration_submissions_answers_check;
ALTER TABLE registration_submissions ADD CONSTRAINT registration_submissions_answers_check CHECK(jsonb_typeof(answers)='array' AND jsonb_array_length(answers)<=40);
