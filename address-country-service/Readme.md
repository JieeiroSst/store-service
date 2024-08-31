```
valueIntrface, err := a.cacheHelper.GetInterface(ctx, common.ListNoteKeyCache, note)
	if err != nil {
		note, errDB = a.noteRepo.NoteAll()
		if errDB != nil {
			log.Error(err.Error())
			return nil, err
		}
		if err == redis.Nil {
			_ = a.cacheHelper.Set(ctx, common.ListNoteKeyCache, note, time.Second*60)
		}
	} else {
		note = valueIntrface.([]model.Note)
	}

```