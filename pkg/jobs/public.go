package jobs

import "context"

func (s *Service) PublicList(ctx context.Context) ([]Job, error) {
	ids := make([]string, 0, 200)
	fetched := make(map[string]*Job, 200)
	for offset := int64(0); len(ids) < 200; offset += 200 {
		batch, err := s.redis.ZRevRange(ctx, s.prefix+"index", offset, offset+199).Result()
		if err != nil {
			return nil, err
		}

		for _, id := range batch {
			job, err := s.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			if job.Private() {
				continue
			}

			fetched[id] = job
			ids = append(ids, id)
			if len(ids) == 200 {
				break
			}
		}

		if len(batch) < 200 {
			break
		}
	}

	ids, err := s.includeActive(ids)
	if err != nil {
		return nil, err
	}

	list := make([]Job, 0, len(ids))
	for _, id := range ids {
		job, ok := fetched[id]
		if !ok {
			job, err = s.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			if job.Private() {
				continue
			}
		}

		clone := *job
		clone.Logs = []string{}
		list = append(list, clone)
	}

	return list, nil
}
