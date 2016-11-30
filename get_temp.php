<?php
//phpinfo();
echo($_REQUEST["temp"]);

if ($_REQUEST["temp"] > 0){
	$open = fopen("/var/www/dash/Nagdash/temp.txt", "w+");
	if ($open){
#	echo($_REQUEST["temp"]);
	fwrite($open, $_REQUEST["temp"]);
	fclose($open);
	}
}
?>
