## Flat Earth

I messed around with TF and gave it a fancy name to make myself feel better about it.

### Schema

For creating a new block, send this sort of request to `/create-new-block`:

```json
{
  "blockType": "resource",
  "blockName": "aws_msk_cluster",
  "attributes": {
    "name": {
      "type": "string",
      "value": "cool-cluster"
    }
  }
}
```
