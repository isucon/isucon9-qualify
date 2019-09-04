<?php


namespace App;

use GuzzleHttp\Client;
use GuzzleHttp\Exception\RequestException;
use PDO;
use Psr\Container\ContainerInterface;
use Psr\Http\Message\UploadedFileInterface;
use Psr\Log\LoggerInterface;
use Slim\Http\Request;
use Slim\Http\Response;
use Slim\Http\StatusCode;

class Service
{
    /**
     * @var LoggerInterface
     */
    private $logger;

    /**
     * @var \PDO
     */
    private $dbh;

    /**
     * @var \SlimSession\Helper
     */
    private $session;

    /**
     * @var array
     */
    private $settings;

    /**
     * @var \Slim\Views\PhpRenderer
     */
    private $renderer;

    private const DATETIME_SQL_FORMAT = 'Y-m-d H:i:s';

    private const ITEM_STATUS_ON_SALE = 'on_sale';
    private const ITEM_STATUS_TRADING = 'trading';
    private const ITEM_STATUS_SOLD_OUT = 'sold_out';
    private const ITEM_STATUS_STOP = 'stop';
    private const ITEM_STATUS_CANCEL = 'cancel';

    private const TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING = 'wait_shipping';
    private const TRANSACTION_EVIDENCE_STATUS_WAIT_DONE = 'wait_done';
    private const TRANSACTION_EVIDENCE_STATUS_DONE = 'done';

    private const SHIPPING_STATUS_INITIAL = 'initial';
    private const SHIPPING_STATUS_WAIT_PICKUP = 'wait_pickup';
    private const SHIPPING_STATUS_SHIPPING = 'shipping';
    private const SHIPPING_STATUS_DONE = 'done';

    private const ISUCARI_API_TOKEN = 'Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0';

    private const PAYMENT_SERVICE_ISUCARI_API_KEY = 'a15400e46c83635eb181-946abb51ff26a868317c';
    private const PAYMENT_SERVICE_ISUCARI_SHOP_ID = '11';

    private const HTTP_USER_AGENT = 'isucon9-qualify-webapp';

    private const MIN_ITEM_PRICE = 100;
    private const MAX_ITEM_PRICE = 1000000;

    private const BUMP_CHARGE_SECONDS = 3;

    private const ITEM_PER_PAGE = 48;
    private const TRANSACTIONS_PER_PAGE = 10;
    private const BCRYPT_COST = 10;

    // constructor receives container instance
    public function __construct(ContainerInterface $container)
    {
        $this->logger = $container->get('logger');
        $this->dbh = $container->get('dbh');
        $this->session = $container->get('session');
        $this->settings = $container->get('settings');
        $this->renderer = $container->get('renderer');
    }

    private function jsonPayload(Request $request)
    {
        $data = json_decode($request->getBody());
        if (JSON_ERROR_NONE !== json_last_error()) {
            throw new \InvalidArgumentException(json_last_error_msg());
        }
        return $data;
    }

    private function getCurrentUser()
    {
        if (! $this->session->exists('user_id')) {
            $this->logger->warning('no session');
            throw new \DomainException('no session');
        }

        $user_id = $this->session->get('user_id');
        $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
        $r = $sth->execute([$user_id]);
        if ($r === false) {
            throw new \PDOException($sth->errorInfo());
        }
        $user = $sth->fetch(PDO::FETCH_ASSOC);

        if ($user === false) {
            $this->logger->warning('not found', ['id' => $user['id']]);
            throw new \DomainException('user not found');
        }

        return $user;
    }

    private function getUserSimpleByID($id)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
        $r = $sth->execute([$id]);
        if ($r === false) {
            throw new \PDOException($sth->errorInfo());
        }
        $user = $sth->fetch(PDO::FETCH_ASSOC);
        if ($user === false) {
            return false;
        }
        return [
          'id' => $user['id'],
            'account_name' => $user['account_name'],
            'num_sell_items' => $user['num_sell_items'],
        ];
    }

    private function simplifyUser($user)
    {
        unset(
            $user['hashed_password'],
            $user['address'],
            $user['last_bump'],
            $user['created_at']
        );
        return $user;
    }

    private function getCategoryByID($id)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `categories` WHERE `id` = ?');
        $r = $sth->execute([$id]);
        if ($r === false) {
            throw new \PDOException($sth->errorInfo());
        }
        $category = $sth->fetch(PDO::FETCH_ASSOC);
        if ($category === false) {
            return false;
        }
        if ((int) $category['parent_id'] !== 0) {
            $parent = $this->getCategoryByID($category['parent_id']);
            if ($parent === false) {
                return false;
            }
            $category['parent_category_name'] = $parent['category_name'];
        }
        return $category;
    }

    private function getImageUrl($name)
    {
        return sprintf("/upload/%s", $name);
    }

    private function getConfigByName($name)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `configs` WHERE `name` = ?');
        $r = $sth->execute([$name]);
        if ($r === false) {
            return "";
        }
        $config = $sth->fetch(PDO::FETCH_ASSOC);
        if ($config === false) {
            return "";
        }
        return $config;
    }

    private function getPaymentServiceURL()
    {
        $config = $this->getConfigByName('payment_service_url');
        if (empty($config['val'])) {
            return "http://localhost:5555";
        }
        return $config['val'];
    }

    private function getShipmentServiceURL()
    {
        $config = $this->getConfigByName('shipment_service_url');
        if (empty($config['val'])) {
            return "http://localhost:7000";
        }
        return $config['val'];
    }

    public function initialize(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        exec($this->settings['app']['base_dir'] . '../sql/init.sh');

        try {
            $sth = $this->dbh->prepare('INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)');
            $r = $sth->execute(["payment_service_url", $payload->payment_service_url]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)');
            $r = $sth->execute(["shipment_service_url", $payload->shipment_service_url]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withJson([
            // キャンペーン実施時には還元率の設定を返す。詳しくはマニュアルを参照のこと。
            "campaign" => 0,
            // 実装言語を返す
            "language" => "php"
        ]);
    }

    public function index(Request $request, Response $response, array $args)
    {
        return $this->renderer->render($response, 'index.html');
    }

    public function new_items(Request $request, Response $response, array $args)
    {
        $itemId = $request->getParam('item_id', "");
        $createdAt = (int) $request->getParam('created_at', 0);

        try {
            if ($itemId !== "" && $createdAt > 0) {
                // paging
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `status` IN (?,?) AND (`created_at` < ? OR (`created_at` <=? AND `id` < ?)) '.
                    'ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $r = $sth->execute([
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_SOLD_OUT,
                    (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                    (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                    $itemId,
                    self::ITEM_PER_PAGE + 1,
                ]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            } else {
                // 1st page
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `status` IN (?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $r = $sth->execute([
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_SOLD_OUT,
                    self::ITEM_PER_PAGE + 1,
                ]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            }
            $items = $sth->fetchAll(PDO::FETCH_ASSOC);

            $itemSimples = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'category not found']);
                }
                $itemSimples[] = [
                  'id' => (int) $item['id'],
                  'seller_id' => (int) $item['seller_id'],
                  'seller' => $seller,
                  'status' => $item['status'],
                  'name' => $item['name'],
                  'price' => (int) $item['price'],
                  'image_url' => $this->getImageUrl($item['image_name']),
                  'category_id' => (int) $item['category_id'],
                  'category' => $category,
                  'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                ];
            }

            $hasNext = false;
            if (count($itemSimples) > self::ITEM_PER_PAGE) {
                $hasNext = true;
                $itemSimples = array_slice($itemSimples, 0, self::ITEM_PER_PAGE);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }
        return $response->withStatus(StatusCode::HTTP_OK)->withJson(
            [
                'items' => $itemSimples,
                'has_next' => $hasNext
            ]
        );
    }

    public function new_category_items(Request $request, Response $response, array $args)
    {
        $rootCategoryId = $args['id'] ?? 0;
        if ((int) $rootCategoryId === 0) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'incorrect category id']);
        }

        $rootCategory = $this->getCategoryByID($rootCategoryId);
        if ($rootCategory === false || (int) $rootCategory['parent_id'] !== 0) {
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson('category not found');
        }

        try {
            $sth = $this->dbh->prepare('SELECT id FROM `categories` WHERE parent_id=?');
            $r = $sth->execute([$rootCategoryId]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $result = $sth->fetchAll(PDO::FETCH_ASSOC);
            $categoryIds = [];
            foreach ($result as $r) {
                $categoryIds[] = $r['id'];
            }

            $itemId = $request->getParam('item_id');
            $createdAt = (int) $request->getParam('created_at', 0);

            if (!empty($itemId) && $createdAt > 0) {
                // paging
                $in = str_repeat('?,', count($categoryIds) - 1) . '?';
                $sth = $this->dbh->prepare("SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (${in}) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ".
                    "ORDER BY `created_at` DESC, `id` DESC LIMIT ?");
                $r = $sth->execute(array_merge(
                    [self::ITEM_STATUS_ON_SALE, self::ITEM_STATUS_SOLD_OUT],
                    $categoryIds,
                    [
                        (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                        (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                        $itemId,
                        self::ITEM_PER_PAGE + 1,
                    ]
                ));
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            } else {
                // 1st page
                $in = str_repeat('?,', count($categoryIds) - 1) . '?';
                $sth = $this->dbh->prepare("SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (${in}) ORDER BY created_at DESC, id DESC LIMIT ?");
                $r = $sth->execute(array_merge(
                    [self::ITEM_STATUS_ON_SALE, self::ITEM_STATUS_SOLD_OUT],
                    $categoryIds,
                    [self::ITEM_PER_PAGE + 1]
                ));
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            }
            $items = $sth->fetchAll(PDO::FETCH_ASSOC);

            $itemSimples = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'category not found']);
                }
                $itemSimples[] = [
                    'id' => $item['id'],
                    'seller_id' => $item['seller_id'],
                    'seller' => $seller,
                    'status' => $item['status'],
                    'name' => $item['name'],
                    'price' => $item['price'],
                    'image_url' => $this->getImageUrl($item['image_name']),
                    'category_id' => $item['category_id'],
                    'category' => $category,
                    'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                ];
            }

            $hasNext = false;
            if (count($itemSimples) > self::ITEM_PER_PAGE) {
                $hasNext = true;
                $itemSimples = array_slice($itemSimples, 0, self::ITEM_PER_PAGE);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson(
            [
                'root_category_id' => (int) $rootCategory['id'],
                'root_category_name' => $rootCategory['category_name'],
                'items' => $itemSimples,
                'has_next' => $hasNext
            ]
        );
    }

    public function user_items(Request $request, Response $response, array $args)
    {
        $userId = $args['id'] ?? 0;

        $user = $this->getUserSimpleByID($userId);
        if ($user === false) {
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        }

        $itemId = $request->getParam('item_id');
        $createdAt = (int) $request->getParam('created_at', 0);
        try {
            if ($itemId !== "" && $createdAt > 0) {
                // paging
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ' .
                            'ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $r = $sth->execute([
                    $user['id'],
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_TRADING,
                    self::ITEM_STATUS_SOLD_OUT,
                    (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                    (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                    $itemId,
                    self::ITEM_PER_PAGE + 1,
                ]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            } else {
                // 1st page
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $r = $sth->execute([
                    $user['id'],
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_TRADING,
                    self::ITEM_STATUS_SOLD_OUT,
                    self::ITEM_PER_PAGE + 1,
                ]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            }
            $items = $sth->fetchAll(PDO::FETCH_ASSOC);

            $itemSimples = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'category not found']);
                }
                $itemSimples[] = [
                    'id' => $item['id'],
                    'seller_id' => $item['seller_id'],
                    'seller' => $seller,
                    'status' => $item['status'],
                    'name' => $item['name'],
                    'price' => $item['price'],
                    'image_url' => $this->getImageUrl($item['image_name']),
                    'category_id' => $item['category_id'],
                    'category' => $category,
                    'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                ];
            }

            $hasNext = false;
            if (count($itemSimples) > self::ITEM_PER_PAGE) {
                $hasNext = true;
                $itemSimples = array_slice($itemSimples, 0, self::ITEM_PER_PAGE);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }
        return $response->withStatus(StatusCode::HTTP_OK)->withJson(
            [
                'user' => $user,
                'items' => $itemSimples,
                'has_next' => $hasNext
            ]
        );
    }

    public function transactions(Request $request, Response $response, array $args)
    {
        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        $itemId = (int) $request->getParam('item_id', 0);
        $createdAt = (int) $request->getParam('created_at', 0);

        try {
            $this->dbh->beginTransaction();

            if ($itemId !== 0 && $createdAt > 0) {
                // paging
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE '.
                    '(`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) AND (`created_at` < ? OR (`created_at` <=? AND `id` < ?)) '.
                    'ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $r = $sth->execute([
                   $user['id'],
                   $user['id'],
                   self::ITEM_STATUS_ON_SALE,
                   self::ITEM_STATUS_TRADING,
                   self::ITEM_STATUS_SOLD_OUT,
                   self::ITEM_STATUS_CANCEL,
                   self::ITEM_STATUS_STOP,
                    (new \DateTime())->setTimeStamp((int) $createdAt)->format(self::DATETIME_SQL_FORMAT),
                    (new \DateTime())->setTimeStamp((int) $createdAt)->format(self::DATETIME_SQL_FORMAT),
                    $itemId,
                    self::TRANSACTIONS_PER_PAGE +1,
                ]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            } else {
                // 1st page
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE ' .
                    '(`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) ' .
                    'ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $r = $sth->execute([
                    $user['id'],
                    $user['id'],
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_TRADING,
                    self::ITEM_STATUS_SOLD_OUT,
                    self::ITEM_STATUS_CANCEL,
                    self::ITEM_STATUS_STOP,
                    self::TRANSACTIONS_PER_PAGE + 1,
                ]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
            }
            $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            $itemDetails = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    $this->dbh->rollBack();
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    $this->dbh->rollBack();
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
                }
                $detail = [
                        'id' => (int) $item['id'],
                        'seller_id' => (int) $item['seller_id'],
                        'seller' => $seller,
                        'status' => $item['status'],
                        'name' => $item['name'],
                        'price' => (int) $item['price'],
                        'description' => $item['description'],
                        'image_url' => $this->getImageUrl($item['image_name']),
                        'category_id' => (int) $item['category_id'],
                        'category' => $category,
                        'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                    ];

                if ((int) $item['buyer_id'] !== 0) {
                    $buyer = $this->getUserSimpleByID($item['buyer_id']);
                    if ($buyer === false) {
                        $this->dbh->rollBack();
                        return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'buyer not found']);
                    }
                    $detail['buyer_id'] = (int) $item['buyer_id'];
                    $detail['buyer'] = $buyer;
                }

                $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
                $r = $sth->execute([$item['id']]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }

                $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
                if ($transactionEvidence !== false) {
                    if ($transactionEvidence['id'] > 0) {
                        $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?');
                        $r = $sth->execute([$transactionEvidence['id']]);
                        if ($r === false) {
                            throw new \PDOException($sth->errorInfo());
                        }
                        $shipping = $sth->fetch(PDO::FETCH_ASSOC);
                        if ($shipping === false) {
                            $this->dbh->rollBack();
                            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'shipping not found']);
                        }

                        $client = new Client();
                        $host = $this->getShipmentServiceURL();
                        try {
                            $r = $client->get($host . '/status', [
                                'headers' => ['Authorization' => self::ISUCARI_API_TOKEN, 'User-Agent' => self::HTTP_USER_AGENT],
                                'json' => ['reserve_id' => $shipping['reserve_id']],
                            ]);
                        } catch (RequestException $e) {
                            $this->dbh->rollBack();
                            if ($e->hasResponse()) {
                                $this->logger->error($e->getResponse()->getReasonPhrase());
                            }
                            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
                        }
                        if ($r->getStatusCode() !== StatusCode::HTTP_OK) {
                            $this->logger->error(($r->getReasonPhrase()));
                            $this->dbh->rollBack();
                            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
                        }
                        $shippingResponse = json_decode($r->getBody());

                        $detail['transaction_evidence_id'] = $transactionEvidence['id'];
                        $detail['transaction_evidence_status'] = $transactionEvidence['status'];
                        $detail['shipping_status'] = $shippingResponse->status;
                    }
                }

                $itemDetails[] = $detail;
            }

            $this->dbh->commit();

            $hasNext = false;
            if (count($itemDetails) > self::TRANSACTIONS_PER_PAGE) {
                $hasNext = true;
                $itemDetails = array_slice($itemDetails, 0, self::TRANSACTIONS_PER_PAGE);
            }
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson([
            'items' => $itemDetails,
            'has_next' => $hasNext,
        ]);
    }

    public function register(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if (empty($payload->account_name) || empty($payload->address) || empty($payload->password)) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'all parameters are required']);
        }

        $hashedPassword = password_hash($payload->password, PASSWORD_BCRYPT, ['cost' => self::BCRYPT_COST]);
        if ($hashedPassword === false) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'error']);
        }

        try {
            $sth = $this->dbh->prepare('INSERT INTO `users` (`account_name`, `hashed_password`, `address`) VALUES (?, ?, ?)');
            $r = $sth->execute([$payload->account_name, $hashedPassword, $payload->address]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $userId = $this->dbh->lastInsertId();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        $this->session->set('user_id', $userId);
        $bytes = random_bytes(20);
        $this->session->set('csrf_token', bin2hex($bytes));

        return $response->withJson(['id' => $userId, 'account_name' => $payload->account_name, 'address' => $payload->address]);
    }

    public function login(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if (empty($payload->account_name) || empty($payload->password)) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'all parameters are required']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `account_name` = ?');
            $r = $sth->execute([$payload->account_name]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $user = $sth->fetch(PDO::FETCH_ASSOC);

            if ($user === false) {
                return $response->withStatus(StatusCode::HTTP_UNAUTHORIZED)->withJson(['error' => 'アカウント名かパスワードが間違えています']);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        if (! password_verify($payload->password, $user['hashed_password'])) {
            return $response->withStatus(StatusCode::HTTP_UNAUTHORIZED)->withJson(['error' => 'アカウント名かパスワードが間違えています']);
        }

        $this->session->set('user_id', $user['id']);
        $bytes = random_bytes(20);
        $this->session->set('csrf_token', bin2hex($bytes));

        return $response->withJson(
            [
                'id' => $user['id'],
                'account_name' => $user['account_name'],
                'address' => $user['address'],
                'num_sell_items' => $user['num_sell_items'],
            ]
        );
    }

    public function settings(Request $request, Response $response, array $args)
    {
        $output = [];
        $output['csrf_token'] = $this->session->get('csrf_token', '');

        try {
            $user = $this->getCurrentUser();
            unset($user['hashed_password'], $user['last_bump'], $user['created_at']);
            $output['user'] = $user;
        } catch (\Exception $e) {
            // pass
        }

        $sth = $this->dbh->query('SELECT * FROM `categories`', PDO::FETCH_ASSOC);
        $sth->execute();
        $categories = $sth->fetchAll();
        if ($categories === false) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }
        $output['categories'] = $categories;
        $output['payment_service_url'] = $this->getPaymentServiceURL();

        return $response->withStatus(StatusCode::HTTP_OK)->withJson($output);
    }

    public function item(Request $request, Response $response, array $args)
    {
        $itemId = $args['id'];

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ?');
            $r = $sth->execute([$itemId]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }
            $item['image_url'] = $this->getImageUrl($item['image_name']);
            $category = $this->getCategoryByID($item['category_id']);
            $item['category'] = $category;

            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
            $r = $sth->execute([$item['seller_id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
            }

            $item['seller'] = $this->simplifyUser($seller);

            if (($user['id'] === $item['seller']['id'] || $user['id'] === $item['buyer_id']) && (int) $item['buyer_id'] !== 0) {
                $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
                $r = $sth->execute([$item['buyer_id']]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
                $buyer = $sth->fetch(PDO::FETCH_ASSOC);
                if ($buyer === false) {
                    return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'buyer not found']);
                }
                $item['buyer'] = $this->simplifyUser($buyer);

                $sth = $this->dbh->prepare("SELECT * FROM `transaction_evidences` WHERE `item_id` = ?");
                $r = $sth->execute([$item['id']]);
                if ($r === false) {
                    throw new \PDOException($sth->errorInfo());
                }
                $transactionEvidence = $sth->fetch();
                if ($transactionEvidence !== false) {
                    $sth = $this->dbh->prepare("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?");
                    $r = $sth->execute([$transactionEvidence["id"]]);
                    if ($r === false) {
                        throw new \PDOException($sth->errorInfo());
                    }
                    $shipping = $sth->fetch();
                    if ($shipping === false) {
                        return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'shipping not found']);
                    }
                    $item['transaction_evidence_id'] = $transactionEvidence["id"];
                    $item['transaction_evidence_status'] = $transactionEvidence["status"];
                    $item['shipping_status'] = $shipping['status'];
                }
            } else {
                unset($item['buyer_id']);
            }
        } catch (\PDOException $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }
        unset($item['updated_at']);
        $item['created_at'] = (new \DateTime($item['created_at']))->getTimestamp();
        return $response->withStatus(StatusCode::HTTP_OK)->withJson($item);
    }


    public function sell(Request $request, Response $response, array $args)
    {
        $csrf_token = $request->getParam('csrf_token', '');
        $name = $request->getParam('name', '');
        $description = $request->getParam('description', '');
        $price = (int) $request->getParam('price', 0);
        $category_id = (int) $request->getParam('category_id', 0);
        /** @var UploadedFileInterface[] $files */
        $files = $request->getUploadedFiles();

        if ($csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        if (empty($name) || empty($description) || empty($price) || $price === 0 || empty($category_id)) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'all parameters are required']);
        }

        if ($price < self::MIN_ITEM_PRICE || $price > self::MAX_ITEM_PRICE) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => '商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください']);
        }

        $category = $this->getCategoryByID($category_id);
        if ($category === false) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'Incorrect category ID']);
        }

        if (! array_key_exists('image', $files)) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'image error']);
        }
        $image = $files['image'];
        $ext = pathinfo($image->getClientFilename(), PATHINFO_EXTENSION);
        if (! in_array($ext, ['jpg', 'jpeg', 'png', 'gif'])) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'unsupported image format error']);
        }
        if ($ext === 'jpeg') {
            $ext = 'jpg';
        }

        $bytes = random_bytes(16);
        $imageName = sprintf("%s.%s", bin2hex($bytes), $ext);
        try {
            $image->moveTo(sprintf('%s/%s', $this->settings['app']['upload_path'], $imageName));
        } catch (\RuntimeException|\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'Saving image failed']);
        }

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $this->dbh->beginTransaction();
            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$user['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                $this->dbh->rollBack();
                $this->logger->warning('seller not found');
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
            }

            $sth = $this->dbh->prepare('INSERT INTO `items` (`seller_id`, `status`, `name`, `price`, `description`, `image_name`, `category_id`) VALUES (?, ?, ?, ?, ?, ?, ?)');
            $r = $sth->execute([
                $seller['id'],
                self::ITEM_STATUS_ON_SALE,
                $name,
                $price,
                $description,
                $imageName,
                $category_id
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $itemId = $this->dbh->lastInsertId();

            $sth = $this->dbh->prepare('UPDATE `users` SET `num_sell_items`=?, `last_bump`=? WHERE `id`=?');
            $r = $sth->execute([
                $seller['num_sell_items']+1,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $seller['id']
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        $this->dbh->commit();

        return $response->withStatus(StatusCode::HTTP_OK)->withJson(['id' => (int) $itemId]);
    }

    public function edit(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        if ($payload->item_price < self::MIN_ITEM_PRICE || $payload->item_price > self::MAX_ITEM_PRICE) {
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => '商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください']);
        }

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ?');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->logger->warning('item not found', ['id' => $payload->item_id]);
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }

            if ($item['seller_id'] !== $user['id']) {
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '自分の商品以外は編集できません']);
            }

            $this->dbh->beginTransaction();
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);

            if ($item['status'] !== self::ITEM_STATUS_ON_SALE) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '販売中の商品以外編集できません']);
            }

            $sth = $this->dbh->prepare('UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?');
            $r = $sth->execute([$payload->item_price, (new \DateTime())->format(self::DATETIME_SQL_FORMAT), $payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ?');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson([
            'item_id' => (int) $item['id'],
            'item_price' => (int) $item['price'],
            'item_created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
            'item_updated_at' => (new \DateTime($item['updated_at']))->getTImestamp(),
        ]);
    }

    public function qrcode(Request $request, Response $response, array $args)
    {
        $transactionEvidenceId = (int) $args['id'];
        try {
            $seller = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ?');
            $r = $sth->execute([$transactionEvidenceId]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['seller_id'] !== $seller['id']) {
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '権限がありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'shippings not found']);
            }

            if ($shipping['status'] !== self::SHIPPING_STATUS_WAIT_PICKUP && $shipping['status'] !== self::SHIPPING_STATUS_SHIPPING) {
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => 'qrcode not available']);
            }

            if (empty($shipping['img_binary'])) {
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'empty qrcode image']);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        $response->getBody()->write($shipping['img_binary']);
        return $response->withHeader('Content-Type', 'image/png');
    }

    public function buy(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        try {
            $buyer = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] !== self::ITEM_STATUS_ON_SALE) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => 'item is not for sale']);
            }

            if ($item['seller_id'] === $buyer['id']) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '自分の商品は買えません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$item['seller_id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'seller not found']);
            }

            $category = $this->getCategoryByID($item['category_id']);
            if ($category === false) {
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'category id error']);
            }

            $sth = $this->dbh->prepare('INSERT INTO `transaction_evidences` '.
                '(`seller_id`, `buyer_id`, `status`, '.
                '`item_id`, `item_name`, `item_price`, `item_description`, '.
                '`item_category_id`, `item_root_category_id`) '.
                'VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)');
            $r = $sth->execute([
                $item['seller_id'],
                $buyer['id'],
                self::TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING,
                $item['id'],
                $item['name'],
                $item['price'],
                $item['description'],
                $category['id'],
                $category['parent_id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidenceId = $this->dbh->lastInsertId();

            $sth = $this->dbh->prepare('UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $r = $sth->execute([
                $buyer['id'],
                self::ITEM_STATUS_TRADING,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $item['id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $client = new Client();
            $host = $this->getShipmentServiceURL();
            try {
                $res = $client->post(
                    $host . '/create',
                    [
                        'headers' => ['Authorization' => self::ISUCARI_API_TOKEN, 'User-Agent' => self::HTTP_USER_AGENT],
                        'json' => [
                            'to_address' => $buyer['address'],
                            'to_name' => $buyer['account_name'],
                            'from_address' => $seller['address'],
                            'from_name' => $seller['account_name'],
                        ]
                    ]
                );
            } catch (RequestException $e) {
                $this->dbh->rollBack();
                if ($e->hasResponse()) {
                    $this->logger->error($e->getResponse()->getReasonPhrase());
                }
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            if ($res->getStatusCode() != StatusCode::HTTP_OK) {
                $this->dbh->rollBack();
                $this->logger->error($res->getReasonPhrase());
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            $shippingResponse = json_decode($res->getBody());

            $host = $this->getPaymentServiceURL();
            try {
                $pres = $client->post(
                    $host . '/token',
                    [
                        'json' => [
                        'shop_id' => self::PAYMENT_SERVICE_ISUCARI_SHOP_ID,
                        'api_key' => self::PAYMENT_SERVICE_ISUCARI_API_KEY,
                        'token' => $payload->token,
                        'price' => $item['price'],
                    ],
                    'headers' => ['User-Agent' => self::HTTP_USER_AGENT],]
                );
            } catch (RequestException $e) {
                $this->dbh->rollBack();
                if ($e->hasResponse()) {
                    $this->logger->error($e->getResponse()->getReasonPhrase());
                }
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'payment service is failed']);
            }

            if ($pres->getStatusCode() != StatusCode::HTTP_OK) {
                $this->dbh->rollBack();
                $this->logger->error($res->getReasonPhrase());
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'payment service is failed']);
            }

            $paymentResponse = json_decode($pres->getBody());
            if (json_last_error() !== JSON_ERROR_NONE) {
                $this->dbh->rollBack();
                $this->logger->error(json_last_error_msg());
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'payment service is failed']);
            }

            if ($paymentResponse->status === 'invalid') {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'カード情報に誤りがあります']);
            }

            if ($paymentResponse->status === 'fail') {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'カードの残高が足りません']);
            }

            if ($paymentResponse->status !== 'ok') {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => '想定外のエラー']);
            }

            $sth = $this->dbh->prepare('INSERT INTO `shippings` '.
                '(`transaction_evidence_id`, `status`, `item_name`, `item_id`, `reserve_id`, `reserve_time`, '.
                '`to_address`, `to_name`, `from_address`, `from_name`, `img_binary`) '.
                'VALUES (?,?,?,?,?,?,?,?,?,?,?)');
            $r = $sth->execute([
                $transactionEvidenceId,
                self::SHIPPING_STATUS_INITIAL,
                $item['name'],
                $item['id'],
                $shippingResponse->reserve_id,
                $shippingResponse->reserve_time,
                $buyer['address'],
                $buyer['account_name'],
                $seller['address'],
                $seller['account_name'],
                "",
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson(['transaction_evidence_id' => (int) $transactionEvidenceId]);
    }

    public function ship(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        try {
            $seller = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['seller_id'] !== $seller['id']) {
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '権限がありません']);
            }

            $this->dbh->beginTransaction();
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] !== self::ITEM_STATUS_TRADING) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => '商品が取引中ではありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['status'] !== self::TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '準備ができていません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'shippings not found']);
            }

            $client = new Client();
            $host = $this->getShipmentServiceURL();
            try {
                $res = $client->post(
                    $host . '/request',
                    [
                        'headers' => ['Authorization' => self::ISUCARI_API_TOKEN, 'User-Agent' => self::HTTP_USER_AGENT],
                        'json' => ['reserve_id' => $shipping['reserve_id']],
                        'stream' => true,
                    ]
                );
            } catch (RequestException $e) {
                $this->dbh->rollBack();
                if ($e->hasResponse()) {
                    $this->logger->error($e->getResponse()->getReasonPhrase());
                }
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            if ($res->getStatusCode() !== StatusCode::HTTP_OK) {
                $this->logger->error($res->getReasonPhrase());
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }

            $sth = $this->dbh->prepare('UPDATE `shippings` SET `status` = ?, `img_binary` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?');
            $r = $sth->execute([
                self::SHIPPING_STATUS_WAIT_PICKUP,
                $res->getBody()->getContents(),
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id']
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson([
            'path' => sprintf("/transactions/%d.png", (int) $transactionEvidence['id']),
            'reserve_id' => (string) $shipping['reserve_id'],
        ]);
    }

    public function ship_done(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        try {
            $seller = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidence not found']);
            }

            if ($transactionEvidence['seller_id'] !== $seller['id']) {
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '権限がありません']);
            }

            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] != self::ITEM_STATUS_TRADING) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '商品が取引中ではありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['status'] !== self::TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '準備ができていません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'shippings not found']);
            }

            $client = new Client();
            $host = $this->getShipmentServiceURL();
            try {
                $r = $client->get($host . '/status', [
                    'headers' => ['Authorization' => self::ISUCARI_API_TOKEN, 'User-Agent' => self::HTTP_USER_AGENT],
                    'json' => ['reserve_id' => $shipping['reserve_id']],
                ]);
            } catch (RequestException $e) {
                $this->dbh->rollBack();
                if ($e->hasResponse()) {
                    $this->logger->error($e->getResponse()->getReasonPhrase());
                }
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            if ($r->getStatusCode() !== StatusCode::HTTP_OK) {
                $this->logger->error($r->getReasonPhrase());
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            $shippingResponse = json_decode($r->getBody());
            if (! ($shippingResponse->status === self::SHIPPING_STATUS_DONE || $shippingResponse->status === self::SHIPPING_STATUS_SHIPPING)) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => 'shipment service側で配送中か配送完了になっていません']);
            }

            $sth = $this->dbh->prepare('UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?');
            $r = $sth->execute([
                $shippingResponse->status,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $r = $sth->execute([
                self::TRANSACTION_EVIDENCE_STATUS_WAIT_DONE,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson(['transaction_evidence_id' => (int) $transactionEvidence['id']]);
    }

    public function complete(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        try {
            $buyer = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidence not found']);
            }

            if ($transactionEvidence['buyer_id'] !== $buyer['id']) {
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '権限がありません']);
            }

            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] != self::ITEM_STATUS_TRADING) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '商品が取引中ではありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['status'] !== self::TRANSACTION_EVIDENCE_STATUS_WAIT_DONE) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '準備ができていません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE');
            $r = $sth->execute([$transactionEvidence['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'shippings not found']);
            }

            $client = new Client();
            $host = $this->getShipmentServiceURL();
            try {
                $r = $client->post($host . '/status', [
                    'headers' => ['Authorization' => self::ISUCARI_API_TOKEN, 'User-Agent' => self::HTTP_USER_AGENT],
                    'json' => ['reserve_id' => $shipping['reserve_id']],
                ]);
            } catch (RequestException $e) {
                $this->dbh->rollBack();
                if ($e->hasResponse()) {
                    $this->logger->error($e->getResponse()->getReasonPhrase());
                }
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            if ($r->getStatusCode() !== StatusCode::HTTP_OK) {
                $this->logger->error($r->getReasonPhrase());
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'failed to request to shipment service']);
            }
            $shippingResponse = json_decode($r->getBody());
            if ($shippingResponse->status !== self::SHIPPING_STATUS_DONE) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'shipment service側で配送完了になっていません']);
            }

            $sth = $this->dbh->prepare('UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?');
            $r = $sth->execute([
                self::SHIPPING_STATUS_DONE,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $r = $sth->execute([
                self::TRANSACTION_EVIDENCE_STATUS_DONE,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('UPDATE `items` SET `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $r = $sth->execute([
                self::ITEM_STATUS_SOLD_OUT,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $item['id'],
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson(['transaction_evidence_id' => (int) $transactionEvidence['id']]);
    }

    public function bump(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_BAD_REQUEST)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(StatusCode::HTTP_UNPROCESSABLE_ENTITY)->withJson(['error' => 'csrf token error']);
        }

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        try {
            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$payload->item_id]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'item not found']);
            }

            if ($item['seller_id'] !== $user['id']) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => '自分の商品以外は編集できません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE');
            $r = $sth->execute([$user['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_NOT_FOUND)->withJson(['error' => 'user not found']);
            }

            // last_bump + 3s > now
            $now = new \DateTime();
            if ((new \DateTime($seller['last_bump']))->getTimestamp() + self::BUMP_CHARGE_SECONDS > $now->getTimestamp()) {
                $this->dbh->rollBack();
                return $response->withStatus(StatusCode::HTTP_FORBIDDEN)->withJson(['error' => 'Bump not allowed']);
            }

            $sth = $this->dbh->prepare('UPDATE `items` SET `created_at`=?, `updated_at`=? WHERE id=?');
            $r = $sth->execute([
                $now->format(self::DATETIME_SQL_FORMAT),
                $now->format(self::DATETIME_SQL_FORMAT),
                $item['id']
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('UPDATE `users` SET `last_bump`=? WHERE id=?');
            $r = $sth->execute([
                $now->format(self::DATETIME_SQL_FORMAT),
                $user['id']
            ]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ?');
            $r = $sth->execute([$item['id']]);
            if ($r === false) {
                throw new \PDOException($sth->errorInfo());
            }
            $item = $sth->fetch(PDO::FETCH_ASSOC);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(StatusCode::HTTP_OK)->withJson([
            'item_id' => (int) $item['id'],
            'item_price' => (int) $item['price'],
            'item_created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
            'item_updated_at' => (new \DateTime($item['updated_at']))->getTimestamp(),
        ]);
    }

    public function reports(Request $request, Response $response, array $args)
    {
        try {
            $sth = $this->dbh->prepare("SELECT * FROM `transaction_evidences` WHERE `id` > 15007");
            $sth->execute([]);
            $transactionEvidences = $sth->fetchAll(PDO::FETCH_ASSOC);
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)->withJson(['error' => 'db error']);
        }

        $t = array_map(function ($e) {
            unset($e['updated_at']);
            unset($e['created_at']);
            return $e;
        }, $transactionEvidences);

        return $response->withJson($t, StatusCode::HTTP_OK);
    }
}
