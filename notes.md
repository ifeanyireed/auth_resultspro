ssh -L 5555:localhost:5555 root@167.99.15.196

//Prisma 
npm install prisma@5 --save-dev
npx prisma init

//launch prisma studio
npx prisma studio --port 5555
//keep ssh terminal open and run these commands in another terminal
ssh -L 5555:localhost:5555 root@167.99.15.196
//then visit this site in another browser http://localhost:5555/

// Build the application
go build -o auth-binary main.go

//seed the database
sqlite3 /var/lib/auth_resultspro/data/auth.db < seed_central_auth.sql

//update the schema
cd /var/www/auth_resultspro
npx prisma generate

//sync the schema with the database
npx prisma db push
nano /var/www/auth_resultspro/prisma/schema.prisma

//make sure the url is this
datasource db {
  provider = "sqlite"
  url      = "file:///var/lib/auth_resultspro/data/auth.db"
}


//users
┌────────────┬─────────────────────────────┬────────────┐
│ Role       │ Email                       │ Password   │
├────────────┼─────────────────────────────┼────────────┤
│ Superadmin │ superadmin@resultspro.ng    │ admin123   │
│ Teacher    │ teacher@example.edu         │ teacher123 │
│ Student    │ student@example.com         │ student123 │
│ Parent     │ parent@example.com          │ parent123  │
│ School Admin│ school-admin@example.edu    │ admin123 │
│ Support Staff│ support-staff@resultspro.ng │ admin123 │
│ Platform Admin│ platform-admin@resultspro.ng│ admin123  │
└────────────┴─────────────────────────────┴────────────┘

//DELETE a user from the database
sqlite3 /var/lib/auth_resultspro/data/auth.db << 'EOF'
DELETE FROM verification_tokens WHERE user_id IN (SELECT id FROM users WHERE email IN ('ifeanyireed@gmail.com', '10myttofficial@gmail.com'));
DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email IN ('ifeanyireed@gmail.com', '10myttofficial@gmail.com'));
DELETE FROM users WHERE email IN ('ifeanyireed@gmail.com', '10myttofficial@gmail.com');
SELECT 'Cleanup complete: Users and associated tokens removed.' as status;
EOF