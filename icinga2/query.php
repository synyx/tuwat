<?php
//hardcode icinga2 api credentials
//our caller should only receive status information w/o credentials nor url params
//this snippet is also to work-around CORS with SSL (javascript requests not in the same domain, or without valid ssl certificate)
#$request_url = "https://infra0.synyx.coffee:5665/v1/status/CIB";
#$request_url = "https://infra0.synyx.coffee:5665/v1/objects/services?joins=host";
$request_url = "https://infra0.synyx.coffee:5665/v1/objects/services";
#$request_url = "https://infra0.synyx.coffee:5665/v1/objects/hosts";
#$request_url = "https://infra0.synyx.coffee:5665/v1/objects/downtimes";

$username = "icingaweb2";
$password = "ra4zahC0Pohce2Ee";
$headers = array(
        'Accept: application/json',
        'X-HTTP-Method-Override: GET'
);
$data = array(
);
$ch = curl_init();
curl_setopt_array($ch, array(
        CURLOPT_URL => $request_url,
        CURLOPT_HTTPHEADER => $headers,
        CURLOPT_USERPWD => $username . ":" . $password,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_CAINFO => "icinga.synyx.coffee.ca.crt", //re-use the icinga2 master ca.crt
        //CURLOPT_POST => count($data),
        //CURLOPT_POSTFIELDS => json_encode($data),
	CURLOPT_SSL_VERIFYHOST => 2,
	CURLOPT_SSL_VERIFYPEER => 1
));
$response = curl_exec($ch);
if ($response === false) {
        print "Error: " . curl_error($ch) . "(" . $response . ")\n";
}
$code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);
print $response;
?>
