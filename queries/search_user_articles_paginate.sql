SELECT uri,
       json_quote(title),
       pubdate
FROM Articles
WHERE title LIKE '%' || :SearchString || '%'
  AND sourceid IN
    (SELECT sourceid
     FROM UserSources
     WHERE userid = :UserID)
ORDER BY pubdate DESC
LIMIT :PageOffset, :PageSize;

