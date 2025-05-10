SELECT uri,
       title,
       pubdate
FROM Articles
WHERE sourceid IN
    (SELECT sourceid
     FROM UserSources
     WHERE userid = :UserID)
ORDER BY pubdate DESC
LIMIT :PageOffset, :PageSize;

