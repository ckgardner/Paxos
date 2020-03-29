package main

//Set method exported
func (replica *Replica) Set(pair Pair, reply *Nothing) error {
	replica.Database[pair.Key] = pair.Value
	return nil
}

// Get method exported
func (replica *Replica) Get(Key string, res *string) error {
	for i := range replica.Database {
		if i == Key {
			*res = replica.Database[i]
		}
	}
	return nil
}

// Delete method exported
func (replica *Replica) Delete(Key string, res *Nothing) error {
	delete(replica.Database, Key)
	return nil
}