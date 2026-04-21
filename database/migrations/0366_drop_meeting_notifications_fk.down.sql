ALTER TABLE meeting_notifications
  ADD CONSTRAINT meeting_notifications_meeting_id_fkey
  FOREIGN KEY (meeting_id)
  REFERENCES microsoft_meetings(id)
  ON DELETE CASCADE;
