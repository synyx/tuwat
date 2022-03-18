<html>
<head>
    <title>Nagios Dashboard</title>
    <meta http-equiv="refresh" content="15">
    <script src="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
    <script src="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/js/bootstrap.min.js"></script>
    <link href="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/css/bootstrap-combined.min.css" rel="stylesheet">
</head>
<link rel="stylesheet" href="css/blinkftw.css">
<link rel="stylesheet" href="css/main.css">
<body>

<?php

function curl_get_file_contents($URL)
{
    $c = curl_init();
    curl_setopt($c, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($c, CURLOPT_URL, $URL);
    $contents = curl_exec($c);
    curl_close($c);

    if ($contents) return $contents;
    else return FALSE;
}

#$ddate = file_get_contents('http://api.ddate.cc/v1/today.txt');
echo "<center><h3>";
if (file_exists('/var/www/dash/Nagdash/ddate.txt')) {
    $ddate = file_get_contents('/var/www/dash/Nagdash/ddate.txt');
    echo "$ddate";
} else {
    if (file_exists('/usr/bin/ddate')) {
        $ddate = exec("/usr/bin/ddate");
        echo "$ddate";
    } else {
        echo "ddate not reachable!";
    }
}
echo "</h3><br><h3>";#
if (file_exists('temp.txt')) {
    include("temp.txt");
    echo "°C";
} else {
    echo "Temperature not reachable!";
}
echo "</h3></center>";

$beer_url = "https://bier.synyx.coffee/dashboard";
$beertime = curl_get_file_contents($beer_url);
if ($beertime != FALSE) {
    if (preg_match('/files\/([^\/]+)\//', file_get_contents($beer_url), $matches) && isset($matches[1])) {
        $biertime = $matches[1] == 'nobeertime' ? 'No Beertime (╯°□°）╯︵ ┻━┻' : 'Beertime ¯\(ツ)/¯';
    } else {
        $biertime = "not sure if its biertime…";
    }
} else {
    $biertime = "biertime not reachable…";
}
echo "<center><h3>$biertime</h3></center><br>";



$nofiles = TRUE;
$of = "done.json";
$op = "progress.json";
$exec_path = dirname(__FILE__);
$CMD = "test.pl";
if ( file_exists($of)) {
  $open = fopen($of, "r");
  $d = json_decode(fread($open, filesize($of)), JSON_PRETTY_PRINT);
  fclose($open);
	$nofiles = FALSE;
} else {
  $d = FALSE;
}
if ( file_exists($of)) {
  $progress = fopen($op, "r");
  $p = json_decode(fread($progress, filesize($op)), JSON_PRETTY_PRINT);
  fclose($progress);
	$nofiles = FALSE;
} else {
  $p = FALSE;
}

if ($nofiles) {
  echo "$of and $op missing :(";
  exec('cd '.$exec_path.' && '.$exec_path.'/'.$CMD.' >>/dev/null 2>&1 &');
} else {
  if ((time()-filemtime($of) > 5 * 60) OR (time()-filemtime($op) > 5 * 60)) {
    exec('cd '.$exec_path.' && '.$exec_path.'/'.$CMD.' >>/dev/null 2>&1 &');
  }
  #echo(var_dump($d));
  $data = "<table align=\"left\"><tr><td colspan=2><b>Done!</b></td></tr>\r\n";

  if (!$d) {
    echo "$of missing :(<br>";
  } else {
    foreach($d as &$gu){

    	if (!is_array($gu)) { continue ; }
      #echo ">>>>".var_dump($gu)."<<<<";
    	#var_dump($gu);
    	foreach ($gu as $lala){
        #var_dump($lala);

        $hero = $lala['assigned_to']['name'];
        if (strlen($hero) == 0){
          $hero = "Shy guy";
        }

        $data .= "<tr><td>".$lala['subject']."</td><td>Hero: <b>". $hero."</b></td></tr>\r\n";

        #var_dump($lala['assigned_to']['name']);
        #

        #die($data);
      }
    }
  }

  if (!$d) {
    echo "$op missing :(<br>";
  } else {
    #echo $data;
    $progress = 0;
    #var_dump($p);
    foreach($p as &$gu){


      #echo ">>>>".var_dump($gu)."<<<<";
    	#var_dump($gu);
    	if (!is_array($gu)) { continue ; }
    	foreach ($gu as $lala){
        #var_dump($lala);

        $hero = $lala['assigned_to']['name'];
        if (strlen($hero) == 0){
          $hero = "Shy guy";
        }
        $progress++;

        $pdata .= "<tr><td>".$lala['subject']."</td><td>Worker-Node: <b>". $hero."</b></td></tr>\r\n";
        #$data .= $lala['subject']."\t\t In Progress by: ". $hero."\r\n";

        #var_dump($lala['assigned_to']['name']);
        #

        #die($data);
      }
    }
    $data .= "<table><tr><td colspan=2><b>In Progress ($progress)~</b></td></tr>\r\n";
    $data .= $pdata;
    echo $data;
    echo "</table>";
  }
}
?>
