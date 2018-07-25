<?php
error_reporting(E_ALL ^ E_NOTICE);
require_once 'config.php';

function connectNagiosApi($url, $action, $payload) {

    switch ($action) {
    case "ack":
        $method = "acknowledge_problem";
        break;
    case "downtime":
        $method = "schedule_downtime";
        $duration = 60 * $_POST['duration'];
        break;
    case "enable":
        $method = "enable_notifications";
        break;
    case "disable":
        $method = "disable_notifications";
        break;
    }

    if (!$method) {
        return "Nagios-api does not support this action ({$action}) yet. ";
    }

    $url .= "/" . $method;

    $params = array('http' =>
        array(
            'method' => 'POST',
            'header' => "Content-type: application/json",
            'content' => json_encode($payload),
        )
    );
    $context = stream_context_create($params);
    if(!$result = file_get_contents($url, false, $context)) {
        $error = error_get_last();
        $error = "Command {$method} failed! <pre>{$error}</pre>";
    } else {
        $return = json_decode($result);
        if (!$return->success) {
            $error = "Command {$method} failed! <pre>{$return->content}</pre>";
            if ($payload['service']) {
                $error .= " -&gt; {$payload['service']}";
            }
        }
    }

    return $error;
}
function connectIcinga2($url, $action, $payload) {

    $username = parse_url($url, PHP_URL_USER);
    $password = parse_url($url, PHP_URL_PASS);

    switch ($action) {
    case "ack":
        $endpoint = 'acknowledge-problem';

        $data = array(
            'author' => $payload['author'],
            'comment' => $payload['comment'],
            'notify' => true,
        );
        break;
    case "downtime":
        $endpoint = 'schedule-downtime';

        $data = array(
            'author' => $payload['author'],
            'comment' => $payload['comment'],
            'start_time' =>  time(),
            'end_time' =>  time() + (60 * $_POST['duration'])
        );
        break;
    case "enable":
        return "Not implemented";
    case "disable":
        return "Not implemented";
    }

    $type = $payload['service'] ? 'Service' : 'Host';
    if ($type == 'Service') {
        $filter = ["servie.name==\"{$payload['host']}!{$payload['service']}\""];
    } else {
        $filter = ["host.name==\"{$payload['host']}\""];
    }
    $filter = implode('&', array_map('urlencode', $filter));
    $request_url = "$url/v1/{$endpoint}?type={$type}&filter=$filter";

    $ch = curl_init();
    curl_setopt_array($ch, array(
        CURLOPT_URL => $request_url,
        CURLOPT_HTTPHEADER => array(
            'Accept: application/json',
            'Content-Type: application/json'
        ),
        CURLOPT_USERPWD => $username . ":" . $password,
        CURLOPT_CUSTOMREQUEST => "POST",
        CURLOPT_POSTFIELDS => json_encode($data),
        CURLOPT_RETURNTRANSFER => true,
        #CURLOPT_CAINFO => "icinga.synyx.coffee.ca.crt", //re-use the icinga2 master ca.crt
        #CURLOPT_SSL_VERIFYHOST => 2,
        #CURLOPT_SSL_VERIFYPEER => 1
        CURLOPT_SSL_VERIFYHOST => 0,
        CURLOPT_SSL_VERIFYPEER => 0
    ));
    if (!$json = curl_exec($ch)) {
        return "<pre>Attempt to hit API failed, sorry. Curl said: " . curl_error($ch) . "</pre>";
    }
    curl_close($ch);

    if (!$state = json_decode($json, true)) {
        return "Attempt to hit API failed, sorry (JSON decode failed)";
    }

    return $error;
}


if (!isset($_POST['nag_host'])) {
    echo "Are you calling this manually? This should be called by Nagdash only.";
} else {
    $nagios_instances = explode(',', $_POST['nag_host']);
    $hostname = $_POST['hostname'];
    # Service is optional
    $service = ($_POST['service']) ? $_POST['service'] : null;
    $action = $_POST['action'];

    $author = function_exists("nagdash_get_user") ? nagdash_get_user() : "Nagdash";

    $payload = array("host" => $hostname, "service" => $service, "comment" => "{$method} from Nagdash", "author" => $author, "duration" => $duration);

    foreach ($nagios_hosts as $host) {
        if (in_array($host['tag'], $nagios_instances)) {
            if ($host['type'] == 'icinga2') {
                $error = connectIcinga2($host['url'], $action, $payload);
            } else {
                $url = $host['protocol'] . "://" . $host['hostname'] . ":" . $host['port'];
                $error = connectNagiosApi($url, $action, $payload);
            }
        }

        if ($error) {
            echo "$error<br/>";
        }
    }

}
