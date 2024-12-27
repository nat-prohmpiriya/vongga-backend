db.auth('admin', 'password')

db = db.getSiblingDB('vongga')

db.createUser({
  user: 'admin',
  pwd: 'password',
  roles: [
    {
      role: 'readWrite',
      db: 'vongga'
    }
  ]
})
