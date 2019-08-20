const USER_ID = 4;
const BASE = 10003;

console.log('use isucari;');
for (let i=0;i<100;i++) {
  console.log(`INSERT INTO items VALUES (${BASE + i},${USER_ID},0,'on_sale','testtest',300,'test_description','sample.jpg',33,'2019-09-07 23:36:47','2019-08-07 23:34:02');`)
}
