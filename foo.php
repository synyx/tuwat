<html>
<head>
<title>Nagios Dashboard</title>
<meta http-equiv="refresh" content="90; URL=http://contargo-nagdash.synyx.coffee/">
<script src="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
<script src="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/js/bootstrap.min.js"></script>
<link href="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/css/bootstrap-combined.min.css" rel="stylesheet">
<link rel="stylesheet" href="css/blinkftw.css">
<link rel="stylesheet" href="css/main.css">

<?php
if (date("G") >= 20){
	echo "<marquee><center><blink><h1>Alles wird gut! &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; echt!</h1></blink></center></marquee>";
}
?>

<?php 
flush();
include("nagdash.php")

##$ddate = file_get_contents('http://api.ddate.cc/v1/today.txt');
#$ddate = file_get_contents('/var/www/dash/Nagdash/ddate.txt');
#echo "<center><h3>$ddate</h3><br>";
#echo "<h3>";
#include("temp.txt");
#echo "°C</h3></center>";
# $biertime = "not sure if its biertime…";
#if (preg_match('/<title>(.+)<\/title>/',file_get_contents('http://bier.synyx.de'),$matches) && isset($matches[1])){
#  $biertime = $matches[1];
#  }
#else{
#     $biertime = "not sure if its biertime…";
#}
##echo "<center><h3>$biertime</h3></center><br>".date("R");
#echo "<center><h3>$biertime</h3></center><br>";

#include("synyx.html");


#flush();
#include("foo2.php");
?>
