-- Retain new columns and historical submissions. Refuse an incompatible downgrade.
DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM event_registration_questions WHERE kind IN ('email','date','time','multiselect','section') OR position>19)
 OR EXISTS(SELECT 1 FROM registration_submissions WHERE jsonb_array_length(answers)>20) THEN
 RAISE EXCEPTION 'Sectioned registrations exist; retain schema 13 and roll back runtime only'; END IF;
END $$;
ALTER TABLE event_registration_questions DROP CONSTRAINT event_registration_questions_kind_check;
ALTER TABLE event_registration_questions ADD CONSTRAINT event_registration_questions_kind_check CHECK(kind IN ('text','textarea','url','number','checkbox','select','phone'));
ALTER TABLE event_registration_questions DROP CONSTRAINT event_registration_questions_position_check;
ALTER TABLE event_registration_questions ADD CONSTRAINT event_registration_questions_position_check CHECK(position BETWEEN 0 AND 19);
ALTER TABLE registration_submissions DROP CONSTRAINT registration_submissions_answers_check;
ALTER TABLE registration_submissions ADD CONSTRAINT registration_submissions_answers_check CHECK(jsonb_typeof(answers)='array' AND jsonb_array_length(answers)<=20);
