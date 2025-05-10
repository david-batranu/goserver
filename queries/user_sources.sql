SELECT uri,
       title,
       id
FROM Sources
where id IN
    (SELECT sourceid
     FROM UserSources
     WHERE userid = ?)
ORDER BY id;

