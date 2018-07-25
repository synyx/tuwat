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

    $url .= "/" + $method;

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

    $error = "not implemented";

    return $error;
}


if (!isset($_POST['nag_host'])) {
    echo "Are you calling this manually? This should be called by Nagdash only.";
} else {
    $nagios_instance = $_POST['nag_host'];
    $hostname = $_POST['hostname'];
    # Service is optional
    $service = ($_POST['service']) ? $_POST['service'] : null;
    $action = $_POST['action'];

    $author = function_exists("nagdash_get_user") ? nagdash_get_user() : "Nagdash";

    if (!$method) {
        echo "Nagios-api does not support this action ({$action}) yet. ";
    } else {
       $payload = array("host" => $hostname, "service" => $service, "comment" => "{$method} from Nagdash", "author" => $author, "duration" => $duration));

        foreach ($nagios_hosts as $host) {
            if ($host['tag'] == $nagios_instance) {
                if ($host['type'] == 'icinga2') {
                    $url = $host['protocol'] . "://" . $host['hostname'] . ":" . $host['port'];
                    $error = connectNagiosApi($url, $action, $payload);
                } else {
                    $error = connectIcinga2($host['url'], $action, $payload);
                }
                break;
            }
        }

        if ($error) {
            echo "$error";
        }
    }
}
