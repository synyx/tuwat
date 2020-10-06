<?php

/**
 * Function that does the dirty to connect to the Nagios API
 *
 * @param string $hostname to connect to.
 * @param int $port to connect to.
 * @param string $protocol for connection.
 * @return string
 */
function connectNagiosApi($hostname, $port, $protocol) {

  global $curl_stats;

  $ch = curl_init("{$protocol}://{$hostname}:{$port}/state");
  curl_setopt($ch, CURLOPT_ENCODING, 'gzip');
  curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
  if (!$json = curl_exec($ch)) {
    return "<pre>Attempt to hit API failed, sorry. Curl said: " . curl_error($ch) . "</pre>";
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  curl_close($ch);

  if (!$state = json_decode($json, true)) {
    return "Attempt to hit API failed, sorry (JSON decode failed)";
  }
  $curl_stats["$hostname:$port"]['objects'] = count($state['content']);
  return $state['content'];
}
